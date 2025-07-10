package subscriptions

type SubscriptionType string

const (
	SubscriptionTypeNext      SubscriptionType = "next"
	SubscriptionTypeComplete  SubscriptionType = "complete"
	SubscriptionTypeSubscribe SubscriptionType = "subscribe"
)
