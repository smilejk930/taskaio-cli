package apiclient

// Meta holds pagination metadata.
type Meta struct {
	NextCursor *string `json:"nextCursor,omitempty"`
	HasMore    bool    `json:"hasMore"`
	Total      *int    `json:"total,omitempty"`
}

// SingleResponse wraps a single resource payload.
type SingleResponse[T any] struct {
	Data T `json:"data"`
}

// ListResponse wraps a list resource payload with cursor pagination metadata.
type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// RawErrorResponse wraps the structured API error response.
type RawErrorResponse struct {
	Error struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// AuthMeResponse defines the data returned by GET /api/v1/auth/me.
type AuthMeResponse struct {
	User struct {
		ID       string  `json:"id"`
		Email    string  `json:"email"`
		Username *string `json:"username,omitempty"`
		IsAdmin  bool    `json:"isAdmin"`
	} `json:"user"`
	Token struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		ExpiresAt *string  `json:"expiresAt,omitempty"`
		Scopes    []string `json:"scopes"`
	} `json:"token"`
}

// Project represents a project resource.
type Project struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatorID   string  `json:"creatorId"`
	Role        *string `json:"role,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// ProjectMember represents a project member resource.
type ProjectMember struct {
	UserID      string  `json:"userId"`
	ProjectID   string  `json:"projectId"`
	Role        string  `json:"role"`
	JoinedAt    string  `json:"joinedAt"`
	DisplayName *string `json:"displayName,omitempty"`
	Username    *string `json:"username,omitempty"`
	Email       *string `json:"email,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	ColorCode   *string `json:"colorCode,omitempty"`
}

// Task represents a task resource.
type Task struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"projectId"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
	Progress    int     `json:"progress"`
	Color       *string `json:"color,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// Schedule represents a holiday/schedule resource.
type Schedule struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Type      string  `json:"type"`
	MemberID  *string `json:"memberId,omitempty"`
	Note      *string `json:"note,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

// CreateProjectInput holds payload for creating a project.
type CreateProjectInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateProjectInput holds payload for updating a project.
type UpdateProjectInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateTaskInput holds payload for creating a task.
type CreateTaskInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
	Progress    *int    `json:"progress,omitempty"`
	Color       *string `json:"color,omitempty"`
}

// UpdateTaskInput holds payload for updating a task.
type UpdateTaskInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	AssigneeID  *string `json:"assigneeId,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
	Progress    *int    `json:"progress,omitempty"`
	Color       *string `json:"color,omitempty"`
}

// CreateScheduleInput holds payload for creating a schedule.
type CreateScheduleInput struct {
	Name      string  `json:"name"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Type      string  `json:"type"`
	MemberID  *string `json:"memberId,omitempty"`
	Note      *string `json:"note,omitempty"`
}

// UpdateScheduleInput holds payload for updating a schedule.
type UpdateScheduleInput struct {
	Name      *string `json:"name,omitempty"`
	StartDate *string `json:"startDate,omitempty"`
	EndDate   *string `json:"endDate,omitempty"`
	Type      *string `json:"type,omitempty"`
	MemberID  *string `json:"memberId,omitempty"`
	Note      *string `json:"note,omitempty"`
}
