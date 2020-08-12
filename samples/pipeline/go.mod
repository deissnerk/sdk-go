module pipeline

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/pipeline => ../../pipeline

replace github.com/cloudevents/sdk-go/pipeline/elements => ../../pipeline/elements

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

require (
	github.com/Azure/go-amqp v0.12.7
	github.com/cloudevents/sdk-go/pipeline v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.1.0
	honnef.co/go/tools v0.0.1-2020.1.4
)
