package environment_logs

// helper function to build a service filter string from provided service ids
func buildServiceFilter(serviceIds []string) string {
	var filterString string

	for i, serviceId := range serviceIds {
		filterString += "@service:" + serviceId
		if i < len(serviceIds)-1 {
			filterString += " OR "
		}
	}

	return filterString
}
