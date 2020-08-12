package resources

//Resource is the interface of a general resource (ComputeInstance,...).
type Resource interface {
	ExtractResource(jsonResourceInfo interface{})
	CompletePricingInfo()
	PrintPricingInfo()
}
