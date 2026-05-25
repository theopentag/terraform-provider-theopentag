package sql

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		NewServerConfigResource,
		NewScheduleResource,
		NewAPIKeyResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerStatusDataSource,
		NewBackupsDataSource,
		NewJobsDataSource,
		NewSSHKeyDataSource,
		NewServersDataSource,
		NewServerConfigsDataSource,
		NewStatsDataSource,
		NewHostStatsDataSource,
		NewPGDatabasesDataSource,
		NewPGUsersDataSource,
		NewUsersDataSource,
	}
}
