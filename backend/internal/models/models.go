package models

import (
	"time"

	"github.com/google/uuid"
)

// ─── USERS ──────────────────────────────────────────────────

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleSeller UserRole = "seller"
	RoleBuyer  UserRole = "buyer"
)

type User struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null"`
	Username      string     `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash  string     `json:"-" gorm:"not null"`
	Role          UserRole   `json:"role" gorm:"type:user_role;default:buyer"`
	IsVerified    bool       `json:"is_verified" gorm:"default:false"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	VerifyToken   *string    `json:"-"`
	ResetToken    *string    `json:"-"`
	ResetExpires  *time.Time `json:"-"`
	GoogleID      *string    `json:"-"`
	LastActive    *time.Time `json:"last_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Profile *Profile `json:"profile,omitempty" gorm:"foreignKey:UserID"`
}

type Profile struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	AvatarURL     string    `json:"avatar_url"`
	CoverURL      string    `json:"cover_url"`
	Bio           string    `json:"bio"`
	Tagline       string    `json:"tagline"`
	Skills        []string  `json:"skills" gorm:"type:text[]"`
	Languages     []string  `json:"languages" gorm:"type:text[]"`
	Location      string    `json:"location"`
	University    string    `json:"university"`
	Department    string    `json:"department"`
	YearOfStudy   int       `json:"year_of_study"`
	PortfolioURL  string    `json:"portfolio_url"`
	GithubURL     string    `json:"github_url"`
	LinkedinURL   string    `json:"linkedin_url"`
	Rating        float64   `json:"rating" gorm:"default:0"`
	TotalReviews  int       `json:"total_reviews" gorm:"default:0"`
	CompletedJobs int       `json:"completed_jobs" gorm:"default:0"`
	ResponseTime  string    `json:"response_time"`
	IsOnline      bool      `json:"is_online" gorm:"default:false"`
	CurrencyPref  string    `json:"currency_pref" gorm:"default:USD"`
	Lat           float64   `json:"lat" gorm:"default:0"`
	Lng           float64   `json:"lng" gorm:"default:0"`
	LocationAt    *time.Time `json:"location_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ─── CATEGORIES ─────────────────────────────────────────────

type Category struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"not null"`
	Slug      string     `json:"slug" gorm:"uniqueIndex;not null"`
	Icon      string     `json:"icon"`
	ParentID  *uuid.UUID `json:"parent_id" gorm:"type:uuid"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
}

// ─── SERVICES ────────────────────────────────────────────────

type Service struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SellerID     uuid.UUID  `json:"seller_id" gorm:"type:uuid;not null"`
	CategoryID   uuid.UUID  `json:"category_id" gorm:"type:uuid;not null"`
	Title        string     `json:"title" gorm:"not null"`
	Slug         string     `json:"slug" gorm:"uniqueIndex;not null"`
	Description  string     `json:"description" gorm:"not null"`
	Tags         []string   `json:"tags" gorm:"type:text[]"`
	Gallery      []string   `json:"gallery" gorm:"type:text[]"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	IsFeatured   bool       `json:"is_featured" gorm:"default:false"`
	Views        int        `json:"views" gorm:"default:0"`
	OrdersCount  int        `json:"orders_count" gorm:"default:0"`
	Rating       float64    `json:"rating" gorm:"default:0"`
	TotalReviews int        `json:"total_reviews" gorm:"default:0"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Seller   *User              `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Category *Category          `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Packages []ServicePackage   `json:"packages,omitempty" gorm:"foreignKey:ServiceID"`
	FAQs     []ServiceFAQ       `json:"faqs,omitempty" gorm:"foreignKey:ServiceID"`
	Reviews  []Review           `json:"reviews,omitempty" gorm:"foreignKey:ServiceID"`
}

type ServicePackage struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceID    uuid.UUID `json:"service_id" gorm:"type:uuid;not null"`
	Name         string    `json:"name" gorm:"not null"` // basic, standard, premium
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description"`
	Price        float64   `json:"price" gorm:"not null"`
	Currency     string    `json:"currency" gorm:"default:USD"`
	DeliveryDays int       `json:"delivery_days" gorm:"not null"`
	Revisions    int       `json:"revisions" gorm:"default:1"`
	Features     []string  `json:"features" gorm:"type:text[]"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
}

type ServiceFAQ struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceID uuid.UUID `json:"service_id" gorm:"type:uuid;not null"`
	Question  string    `json:"question" gorm:"not null"`
	Answer    string    `json:"answer" gorm:"not null"`
	SortOrder int       `json:"sort_order"`
}

// ─── ORDERS ──────────────────────────────────────────────────

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderInProgress OrderStatus = "in_progress"
	OrderDelivered  OrderStatus = "delivered"
	OrderRevision   OrderStatus = "revision"
	OrderCompleted  OrderStatus = "completed"
	OrderCancelled  OrderStatus = "cancelled"
	OrderDisputed   OrderStatus = "disputed"
)

type Order struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BuyerID        uuid.UUID   `json:"buyer_id" gorm:"type:uuid;not null"`
	SellerID       uuid.UUID   `json:"seller_id" gorm:"type:uuid;not null"`
	ServiceID      uuid.UUID   `json:"service_id" gorm:"type:uuid;not null"`
	PackageID      uuid.UUID   `json:"package_id" gorm:"type:uuid;not null"`
	Status         OrderStatus `json:"status" gorm:"type:order_status;default:pending"`
	Amount         float64     `json:"amount" gorm:"not null"`
	PlatformFee    float64     `json:"platform_fee" gorm:"not null"`
	SellerAmount   float64     `json:"seller_amount" gorm:"not null"`
	Currency       string      `json:"currency" gorm:"default:USD"`
	Requirements   string      `json:"requirements"`
	DeliveryFiles  []string    `json:"delivery_files" gorm:"type:text[]"`
	RevisionCount  int         `json:"revision_count" gorm:"default:0"`
	MaxRevisions   int         `json:"max_revisions" gorm:"default:1"`
	DueDate        *time.Time  `json:"due_date"`
	DeliveredAt    *time.Time  `json:"delivered_at"`
	CompletedAt    *time.Time  `json:"completed_at"`
	CancelledAt    *time.Time  `json:"cancelled_at"`
	CancelReason   string      `json:"cancel_reason"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`

	Buyer   *User           `json:"buyer,omitempty" gorm:"foreignKey:BuyerID"`
	Seller  *User           `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Service *Service        `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
	Package *ServicePackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
	Review  *Review         `json:"review,omitempty" gorm:"foreignKey:OrderID"`
}

// ─── CHAT ────────────────────────────────────────────────────

type Chat struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID   *uuid.UUID `json:"order_id" gorm:"type:uuid"`
	IsGroup   bool      `json:"is_group" gorm:"default:false"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Participants []ChatParticipant `json:"participants,omitempty" gorm:"foreignKey:ChatID"`
	Messages     []ChatMessage     `json:"messages,omitempty" gorm:"foreignKey:ChatID"`
	LastMessage  *ChatMessage      `json:"last_message,omitempty" gorm:"-"`
	UnreadCount  int               `json:"unread_count,omitempty" gorm:"-"`
}

type ChatParticipant struct {
	ChatID   uuid.UUID  `json:"chat_id" gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID  `json:"user_id" gorm:"type:uuid;primaryKey"`
	JoinedAt time.Time  `json:"joined_at"`
	LastRead *time.Time `json:"last_read"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChatID    uuid.UUID `json:"chat_id" gorm:"type:uuid;not null"`
	SenderID  uuid.UUID `json:"sender_id" gorm:"type:uuid;not null"`
	Content   string    `json:"content"`
	FileURL   string    `json:"file_url"`
	FileName  string    `json:"file_name"`
	FileType  string    `json:"file_type"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	IsDeleted bool      `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`

	Sender *User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

// ─── POSTS ───────────────────────────────────────────────────

type Post struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AuthorID      uuid.UUID `json:"author_id" gorm:"type:uuid;not null"`
	Content       string    `json:"content" gorm:"not null"`
	Images        []string  `json:"images" gorm:"type:text[]"`
	LikesCount    int       `json:"likes_count" gorm:"default:0"`
	CommentsCount int       `json:"comments_count" gorm:"default:0"`
	SharesCount   int       `json:"shares_count" gorm:"default:0"`
	IsPinned      bool      `json:"is_pinned" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Author   *User     `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:PostID"`
	IsLiked  bool      `json:"is_liked,omitempty" gorm:"-"`
}

type Comment struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PostID     uuid.UUID  `json:"post_id" gorm:"type:uuid;not null"`
	AuthorID   uuid.UUID  `json:"author_id" gorm:"type:uuid;not null"`
	ParentID   *uuid.UUID `json:"parent_id" gorm:"type:uuid"`
	Content    string     `json:"content" gorm:"not null"`
	LikesCount int        `json:"likes_count" gorm:"default:0"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	Author  *User     `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Replies []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

// ─── REVIEWS ─────────────────────────────────────────────────

type Review struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID        uuid.UUID  `json:"order_id" gorm:"type:uuid;uniqueIndex;not null"`
	ServiceID      uuid.UUID  `json:"service_id" gorm:"type:uuid;not null"`
	ReviewerID     uuid.UUID  `json:"reviewer_id" gorm:"type:uuid;not null"`
	SellerID       uuid.UUID  `json:"seller_id" gorm:"type:uuid;not null"`
	Rating         int        `json:"rating" gorm:"not null"`
	Content        string     `json:"content"`
	SellerResponse string     `json:"seller_response"`
	RespondedAt    *time.Time `json:"responded_at"`
	CreatedAt      time.Time  `json:"created_at"`

	Reviewer *User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
}

// ─── NOTIFICATIONS ───────────────────────────────────────────

type NotificationType string

const (
	NotifOrder         NotificationType = "order"
	NotifMessage       NotificationType = "message"
	NotifReview        NotificationType = "review"
	NotifFollow        NotificationType = "follow"
	NotifLike          NotificationType = "like"
	NotifComment       NotificationType = "comment"
	NotifSystem        NotificationType = "system"
	NotifFriendRequest NotificationType = "friend_request"
)

type Notification struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID        `json:"user_id" gorm:"type:uuid;not null"`
	Type      NotificationType `json:"type" gorm:"type:notification_type;not null"`
	Title     string           `json:"title" gorm:"not null"`
	Body      string           `json:"body" gorm:"not null"`
	Data      map[string]any   `json:"data,omitempty" gorm:"type:jsonb;serializer:json"`
	IsRead    bool             `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time        `json:"created_at"`
}

// ─── FRIENDSHIP ──────────────────────────────────────────────

type FriendshipStatus string

const (
	FriendPending  FriendshipStatus = "pending"
	FriendAccepted FriendshipStatus = "accepted"
	FriendBlocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID        `json:"requester_id" gorm:"type:uuid;not null"`
	AddresseeID uuid.UUID        `json:"addressee_id" gorm:"type:uuid;not null"`
	Status      FriendshipStatus `json:"status" gorm:"type:friendship_status;default:pending"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	Requester *User `json:"requester,omitempty" gorm:"foreignKey:RequesterID"`
	Addressee *User `json:"addressee,omitempty" gorm:"foreignKey:AddresseeID"`
}

type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id" gorm:"type:uuid;primaryKey"`
	FollowingID uuid.UUID `json:"following_id" gorm:"type:uuid;primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── LIKES ───────────────────────────────────────────────────

type Like struct {
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;primaryKey"`
	PostID    *uuid.UUID `json:"post_id,omitempty" gorm:"type:uuid"`
	CommentID *uuid.UUID `json:"comment_id,omitempty" gorm:"type:uuid"`
	CreatedAt time.Time  `json:"created_at"`
}

// ─── REPORTS ─────────────────────────────────────────────────

type Report struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReporterID     uuid.UUID  `json:"reporter_id" gorm:"type:uuid;not null"`
	ReportedUserID *uuid.UUID `json:"reported_user_id" gorm:"type:uuid"`
	ServiceID      *uuid.UUID `json:"service_id" gorm:"type:uuid"`
	PostID         *uuid.UUID `json:"post_id" gorm:"type:uuid"`
	Reason         string     `json:"reason" gorm:"not null"`
	Description    string     `json:"description"`
	Status         string     `json:"status" gorm:"default:pending"`
	AdminNote      string     `json:"admin_note"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at"`

	Reporter *User `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
}

// ─── ADMIN ───────────────────────────────────────────────────

type AdminLog struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AdminID    uuid.UUID      `json:"admin_id" gorm:"type:uuid;not null"`
	Action     string         `json:"action" gorm:"not null"`
	TargetType string         `json:"target_type"`
	TargetID   *uuid.UUID     `json:"target_id" gorm:"type:uuid"`
	Metadata   map[string]any `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json"`
	IPAddress  string         `json:"ip_address"`
	CreatedAt  time.Time      `json:"created_at"`
}
