package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"infra/config"
)

type InfraStackProps struct {
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, id string, props *InfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("InfraQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	// Create role for lambda function.
	lambdaRole := awsiam.NewRole(stack, jsii.String("LambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(*stack.StackName() + "-LambdaRole"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
		},
	})

	// Create cat facts function.
	catFactFunction := awslambda.NewFunction(stack, jsii.String("GetCatFacts"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-GetCatFacts"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../out/."), nil),
		Handler:      jsii.String("retrieve_catfact_data_linux"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		CurrentVersionOptions: &awslambda.VersionOptions{
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
		},
	})

	// Create item function
	createItemFunction := awslambda.NewFunction(stack, jsii.String("CreateItem"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-CreateItem"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../out/."), nil),
		Handler:      jsii.String("create_item_entry_linux"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_FIVE_DAYS,
		CurrentVersionOptions: &awslambda.VersionOptions{
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
		},
	})

	// Scan items function
	scanItemsFunction := awslambda.NewFunction(stack, jsii.String("ScanItems"), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-ScanItems"),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Code:         awslambda.AssetCode_FromAsset(jsii.String("../out/."), nil),
		Handler:      jsii.String("scan_items_linux"),
		Architecture: awslambda.Architecture_X86_64(),
		Role:         lambdaRole,
		LogRetention: awslogs.RetentionDays_FIVE_DAYS,
		CurrentVersionOptions: &awslambda.VersionOptions{
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		},
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": jsii.String(*stack.StackName() + "-" + config.DynamoDBTable),
		},
	})

	// API Gateway Configuration
	// Create API Gateway rest api.
	restApi := awsapigateway.NewRestApi(stack, jsii.String("LambdaRestApi"), &awsapigateway.RestApiProps{
		RestApiName:        jsii.String(*stack.StackName() + "-LambdaRestApi"),
		RetainDeployments:  jsii.Bool(false),
		EndpointExportName: jsii.String("RestApiUrl"),
		Deploy:             jsii.Bool(true),
		EndpointConfiguration: &awsapigateway.EndpointConfiguration{
			Types: &[]awsapigateway.EndpointType{
				awsapigateway.EndpointType_REGIONAL,
			},
		},
		DeployOptions: &awsapigateway.StageOptions{
			StageName:           jsii.String("dev"),
			CacheClusterEnabled: jsii.Bool(false),
			CacheTtl:            awscdk.Duration_Minutes(jsii.Number(1)),
			// https://www.petefreitag.com/item/853.cfm
			// This can help you better understand what burst and rate limit are.
			ThrottlingBurstLimit: jsii.Number(100),
			ThrottlingRateLimit:  jsii.Number(1000),
		},
	})

	// Cat fact endpoint
	getCatFactRes := restApi.Root().AddResource(jsii.String("get-cat-fact"), nil)
	getCatFactRes.AddMethod(jsii.String("GET"), awsapigateway.NewLambdaIntegration(catFactFunction, nil),
		&awsapigateway.MethodOptions{
			ApiKeyRequired: jsii.Bool(true),
		})

	// Create Item endpoint
	addItemResource := restApi.Root().AddResource(jsii.String("add-item"), nil)
	addItemResource.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(createItemFunction, nil),
		&awsapigateway.MethodOptions{
			ApiKeyRequired: jsii.Bool(true),
		})

	// Scan items endpoint
	scanItemsResource := restApi.Root().AddResource(jsii.String("scan-items"), nil)
	scanItemsResource.AddMethod(jsii.String("GET"), awsapigateway.NewLambdaIntegration(scanItemsFunction, nil),
		&awsapigateway.MethodOptions{
			ApiKeyRequired: jsii.Bool(true),
		})

	// UsagePlane's throttle can override Stage's DefaultMethodThrottle,
	// while UsagePlanePerApiStage's throttle can override UsagePlane's throttle.
	usagePlane := restApi.AddUsagePlan(jsii.String("UsagePlane"), &awsapigateway.UsagePlanProps{
		Name: jsii.String(*stack.StackName() + "-UsagePlane"),
		Throttle: &awsapigateway.ThrottleSettings{
			BurstLimit: jsii.Number(10),
			RateLimit:  jsii.Number(100),
		},
		Quota: &awsapigateway.QuotaSettings{
			Limit:  jsii.Number(100),
			Offset: jsii.Number(0),
			Period: awsapigateway.Period_DAY,
		},
		ApiStages: &[]*awsapigateway.UsagePlanPerApiStage{
			{
				Api:      restApi,
				Stage:    restApi.DeploymentStage(),
				Throttle: &[]*awsapigateway.ThrottlingPerMethod{
					/*										{
															Method: getMethod,
															Throttle: &awsapigateway.ThrottleSettings{
																BurstLimit: jsii.Number(1),
																RateLimit:  jsii.Number(10),
															},
														},*/
				},
			},
		},
	})

	// Create ApiKey and associate it with UsagePlane.
	apiKey := restApi.AddApiKey(jsii.String("ApiKey"), &awsapigateway.ApiKeyOptions{})
	usagePlane.AddApiKey(apiKey, &awsapigateway.AddApiKeyOptions{})

	// Create DynamoDB Base table.
	// Data Modeling
	// name(PK), time(SK),                  comment, chat_room
	// string    string(micro sec unixtime)	string   string
	itemTable := awsdynamodb.NewTable(stack, jsii.String(config.DynamoDBTable), &awsdynamodb.TableProps{
		TableName:     jsii.String("FoodItems"),
		BillingMode:   awsdynamodb.BillingMode_PROVISIONED,
		ReadCapacity:  jsii.Number(1),
		WriteCapacity: jsii.Number(1),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("itemId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("storageLocation"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		PointInTimeRecovery: jsii.Bool(true),
	})

	itemTable.AutoScaleWriteCapacity(&awsdynamodb.EnableScalingProps{
		MinCapacity: jsii.Number(1),
		MaxCapacity: jsii.Number(5),
	})

	itemTable.AutoScaleReadCapacity(&awsdynamodb.EnableScalingProps{
		MinCapacity: jsii.Number(1),
		MaxCapacity: jsii.Number(5),
	})

	// Create DynamoDB GSI table.
	// Data Modeling
	// chat_room(PK), time(SK),                  comment, name
	// string         string(micro sec unixtime) string   string
	/*	itemTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String(config.DynamoDBGSI),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("chat_room"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("time"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})*/

	// Grant access to lambda functions.
	itemTable.GrantWriteData(createItemFunction)
	//itemTable.GrantReadData(getFunction)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewInfraStack(app, "FreezerStack", &InfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
