package stripe

import (
	"fmt"

	"github.com/quiver/billing/pkg/models"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/customer"
	"github.com/stripe/stripe-go/v75/price"
	"github.com/stripe/stripe-go/v75/product"
	"github.com/stripe/stripe-go/v75/subscription"
	"github.com/stripe/stripe-go/v75/webhook"
)

// Client wraps Stripe API operations
type Client struct {
	apiKey          string
	webhookSecret   string
	priceIDs        map[models.PlanType]string
}

// NewClient creates a new Stripe client
func NewClient(apiKey, webhookSecret string) *Client {
	stripe.Key = apiKey
	
	return &Client{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		priceIDs: map[models.PlanType]string{
			models.PlanStarter:    "", // Set after creating products
			models.PlanPro:        "",
			models.PlanEnterprise: "",
		},
	}
}

// CreateProducts creates Stripe products and prices for each plan
func (c *Client) CreateProducts() error {
	plans := []struct {
		plan        models.PlanType
		name        string
		description string
		price       int64
	}{
		{
			plan:        models.PlanStarter,
			name:        "QUIVer Starter",
			description: "10,000 requests/month, up to 7B models",
			price:       999,
		},
		{
			plan:        models.PlanPro,
			name:        "QUIVer Pro",
			description: "100,000 requests/month, all models",
			price:       4999,
		},
		{
			plan:        models.PlanEnterprise,
			name:        "QUIVer Enterprise",
			description: "Unlimited requests, custom models, SLA",
			price:       29999,
		},
	}

	for _, p := range plans {
		// Create product
		productParams := &stripe.ProductParams{
			Name:        stripe.String(p.name),
			Description: stripe.String(p.description),
		}
		prod, err := product.New(productParams)
		if err != nil {
			return fmt.Errorf("failed to create product %s: %w", p.name, err)
		}

		// Create price
		priceParams := &stripe.PriceParams{
			Product:    stripe.String(prod.ID),
			UnitAmount: stripe.Int64(p.price),
			Currency:   stripe.String("usd"),
			Recurring: &stripe.PriceRecurringParams{
				Interval: stripe.String("month"),
			},
		}
		pr, err := price.New(priceParams)
		if err != nil {
			return fmt.Errorf("failed to create price for %s: %w", p.name, err)
		}

		c.priceIDs[p.plan] = pr.ID
	}

	return nil
}

// CreateCustomer creates a Stripe customer
func (c *Client) CreateCustomer(email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"platform": "quiver",
		},
	}

	return customer.New(params)
}

// CreateSubscription creates a subscription for a customer
func (c *Client) CreateSubscription(customerID string, plan models.PlanType) (*stripe.Subscription, error) {
	priceID, ok := c.priceIDs[plan]
	if !ok {
		return nil, fmt.Errorf("no price ID for plan %s", plan)
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		Expand:          []*string{stripe.String("latest_invoice.payment_intent")},
	}

	return subscription.New(params)
}

// UpdateSubscription updates a subscription to a new plan
func (c *Client) UpdateSubscription(subscriptionID string, newPlan models.PlanType) (*stripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	priceID, ok := c.priceIDs[newPlan]
	if !ok {
		return nil, fmt.Errorf("no price ID for plan %s", newPlan)
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(priceID),
			},
		},
	}

	return subscription.Update(subscriptionID, params)
}

// CancelSubscription cancels a subscription at period end
func (c *Client) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	return subscription.Update(subscriptionID, params)
}

// HandleWebhook processes Stripe webhooks
func (c *Client) HandleWebhook(payload []byte, signature string) (*stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, c.webhookSecret)
}

// CreateUsageRecord records usage for metered billing
func (c *Client) CreateUsageRecord(subscriptionItemID string, quantity int64) error {
	// For usage-based billing in enterprise plans
	params := &stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(subscriptionItemID),
		Quantity:         stripe.Int64(quantity),
		Timestamp:        stripe.Int64(0), // Current time
		Action:           stripe.String("increment"),
	}
	
	_, err := subscription.NewUsageRecord(params)
	return err
}

// GetPriceID returns the Stripe price ID for a plan
func (c *Client) GetPriceID(plan models.PlanType) string {
	return c.priceIDs[plan]
}