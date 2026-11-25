package client

type MeshUser struct {
	Metadata MeshUserMetadata `json:"metadata" tfsdk:"metadata"`
	Spec     MeshUserSpec     `json:"spec" tfsdk:"spec"`
	Status   MeshUserStatus   `json:"status" tfsdk:"status"`
}

type MeshUserMetadata struct {
	Name string `json:"name" tfsdk:"name"`
	Uuid string `json:"uuid" tfsdk:"uuid"`
}

type MeshUserSpec struct {
	Email     string              `json:"email" tfsdk:"email"`
	FirstName string              `json:"firstName" tfsdk:"first_name"`
	LastName  string              `json:"lastName" tfsdk:"last_name"`
	Euid      *string             `json:"euid,omitzero" tfsdk:"euid"`
	Tags      map[string][]string `json:"tags" tfsdk:"tags"`
}

type MeshUserStatus struct {
	Lifecycle MeshUserLifecycle `json:"lifecycle" tfsdk:"lifecycle"`
	LastLogin *string           `json:"lastLogin,omitzero" tfsdk:"last_login"`
}

type MeshUserLifecycle struct {
	State string `json:"state" tfsdk:"state"`
}

type MeshUserFilter struct {
	Uuid                string `json:"uuid"`
	Username            string `json:"username"`
	Email               string `json:"email"`
	Euid                string `json:"euid"`
	WorkspaceIdentifier string `json:"workspaceIdentifier"`
	WorkspaceRole       string `json:"workspaceRole"`
	ProjectIdentifier   string `json:"projectIdentifier"`
}

type MeshUserClient = MeshObjectClient[MeshUser, MeshUserFilter]

func (user MeshUser) meshObjectId() string {
	return user.Metadata.Name
}
