package railway

// yucky
var projectQuery = `query project($id: String!) {
		project(id: $id) {
		  id
		  name
		  description
		  environments {
			edges {
			  node {
				id
				name
			  }
			}
		  }
		  services {
			edges {
			  node {
				id
				name
				serviceInstances {
				  edges {
					node {
					  environmentId
					}
				  }
				}
			  }
			}
		  }
		}
	  }`

var environmentQuery = `query environment($id: String!) {
		environment(id: $id) {
		  projectId
		}
	  }`

var streamEnvironmentLogsQuery = `subscription streamEnvironmentLogs(
		$environmentId: String!
		$filter: String
		$beforeLimit: Int!
		$beforeDate: String
		$anchorDate: String
		$afterDate: String
		$afterLimit: Int
	  ) {
		environmentLogs(
		  environmentId: $environmentId
		  filter: $filter
		  beforeDate: $beforeDate
		  anchorDate: $anchorDate
		  afterDate: $afterDate
		  beforeLimit: $beforeLimit
		  afterLimit: $afterLimit
		) {
		  timestamp
		  message
		  severity
		  tags {
			projectId
			environmentId
			pluginId
			serviceId
			deploymentId
			deploymentInstanceId
			snapshotId
		  }
		  attributes {
			key
			value
		  }
		}
	  }`
