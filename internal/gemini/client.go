package gemini

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/holycann/cultour-backend/configs"
	achievementModels "github.com/holycann/cultour-backend/internal/achievement/models"
	culturalModels "github.com/holycann/cultour-backend/internal/cultural/models"
	discussionModels "github.com/holycann/cultour-backend/internal/discussion/models"
	placeModels "github.com/holycann/cultour-backend/internal/place/models"
	userModels "github.com/holycann/cultour-backend/internal/users/models"
	"golang.org/x/time/rate"
	"google.golang.org/genai"
)

// SafetyFilter manages content safety and quality
type SafetyFilter struct {
	MinResponseQuality float32
	MaxResponseLength  int
}

// ValidateResponseQuality checks the quality of the AI response
func (sf *SafetyFilter) ValidateResponseQuality(response string) bool {
	// Check response length
	if len(response) > sf.MaxResponseLength {
		return false
	}

	// Basic quality assessment
	words := strings.Fields(response)
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[word] = true
	}

	// Calculate quality score
	uniqueWordRatio := float32(len(uniqueWords)) / float32(len(words)+1)
	lengthFactor := float32(len(response)) / 1000.0 // Normalize to 0-1

	qualityScore := (uniqueWordRatio + lengthFactor) / 2

	return qualityScore >= sf.MinResponseQuality
}

// ComprehensiveKnowledgeBase integrates multiple domain models
type ComprehensiveKnowledgeBase struct {
	mutex sync.RWMutex

	// User-related knowledge
	users        map[string]*userModels.User
	userProfiles map[string]*userModels.UserProfile
	userBadges   map[string][]*userModels.UserBadge

	// Cultural domain
	events       map[string]*culturalModels.Event
	localStories map[string]*culturalModels.LocalStory

	// Place domain
	cities    map[string]*placeModels.City
	provinces map[string]*placeModels.Province
	locations map[string]*placeModels.Location

	// Discussion domain
	threads  map[string]*discussionModels.Thread
	messages map[string]*discussionModels.Message

	// Achievement domain
	badges map[string]*achievementModels.Badge

	// Contextual metadata
	contextualFacts map[string]string
}

// NewComprehensiveKnowledgeBase initializes an integrated knowledge base
func NewComprehensiveKnowledgeBase() *ComprehensiveKnowledgeBase {
	return &ComprehensiveKnowledgeBase{
		users:           make(map[string]*userModels.User),
		userProfiles:    make(map[string]*userModels.UserProfile),
		userBadges:      make(map[string][]*userModels.UserBadge),
		events:          make(map[string]*culturalModels.Event),
		localStories:    make(map[string]*culturalModels.LocalStory),
		cities:          make(map[string]*placeModels.City),
		provinces:       make(map[string]*placeModels.Province),
		locations:       make(map[string]*placeModels.Location),
		threads:         make(map[string]*discussionModels.Thread),
		messages:        make(map[string]*discussionModels.Message),
		badges:          make(map[string]*achievementModels.Badge),
		contextualFacts: make(map[string]string),
	}
}

// Methods for adding and retrieving knowledge across domains

// AddUser adds or updates user information
func (kb *ComprehensiveKnowledgeBase) AddUser(user *userModels.User) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.users[user.ID] = user
}

// AddUserProfile adds or updates user profile
func (kb *ComprehensiveKnowledgeBase) AddUserProfile(profile *userModels.UserProfile) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.userProfiles[profile.UserID.String()] = profile
}

// AddUserBadge adds a badge to a user's collection
func (kb *ComprehensiveKnowledgeBase) AddUserBadge(userID string, userBadge *userModels.UserBadge) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.userBadges[userID] = append(kb.userBadges[userID], userBadge)
}

// AddEvent adds or updates an event
func (kb *ComprehensiveKnowledgeBase) AddEvent(event *culturalModels.Event) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.events[event.ID.String()] = event
}

// AddLocalStory adds or updates a local story
func (kb *ComprehensiveKnowledgeBase) AddLocalStory(story *culturalModels.LocalStory) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.localStories[story.ID.String()] = story
}

// AddCity adds or updates a city
func (kb *ComprehensiveKnowledgeBase) AddCity(city *placeModels.City) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.cities[city.ID.String()] = city
}

// AddProvince adds or updates a province
func (kb *ComprehensiveKnowledgeBase) AddProvince(province *placeModels.Province) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.provinces[province.ID.String()] = province
}

// AddLocation adds or updates a location
func (kb *ComprehensiveKnowledgeBase) AddLocation(location *placeModels.Location) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.locations[location.ID.String()] = location
}

// AddThread adds or updates a discussion thread
func (kb *ComprehensiveKnowledgeBase) AddThread(thread *discussionModels.Thread) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.threads[thread.ID.String()] = thread
}

// AddMessage adds or updates a message
func (kb *ComprehensiveKnowledgeBase) AddMessage(message *discussionModels.Message) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.messages[message.ID.String()] = message
}

// AddBadge adds or updates a badge
func (kb *ComprehensiveKnowledgeBase) AddBadge(badge *achievementModels.Badge) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.badges[badge.ID.String()] = badge
}

// AddContextualFact adds a general contextual fact
func (kb *ComprehensiveKnowledgeBase) AddContextualFact(key, fact string) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.contextualFacts[key] = fact
}

// BuildContextualPrompt generates a comprehensive context for AI interactions
func (kb *ComprehensiveKnowledgeBase) BuildContextualPrompt(userID string, eventID *string) string {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	// Start with base context from system policies
	contextParts := []string{
		GetSystemPolicies(System, Behavior, Feature),
		"Provide a response that strictly adheres to the application's cultural tourism scope.",
	}

	// Add user context if available
	if profile, exists := kb.userProfiles[userID]; exists {
		contextParts = append(contextParts,
			fmt.Sprintf("User Context: Name - %s", profile.Fullname),
		)
	}

	// Add user profile context
	if profile, exists := kb.userProfiles[userID]; exists {
		contextParts = append(contextParts,
			fmt.Sprintf("User Profile: UserID - %s", profile.UserID),
		)
	}

	// Add event context if specified
	if eventID != nil {
		if event, exists := kb.events[*eventID]; exists {
			contextParts = append(contextParts,
				fmt.Sprintf("Event Context: %s from %s to %s",
					event.Name,
					event.StartDate.Format("2006-01-02"),
					event.EndDate.Format("2006-01-02"),
				),
			)
		}
	}

	// Add user badges context
	if userBadges, exists := kb.userBadges[userID]; exists {
		badgeNames := make([]string, len(userBadges))
		for i, badge := range userBadges {
			badgeNames[i] = badge.BadgeID.String()
		}
		contextParts = append(contextParts,
			fmt.Sprintf("User Badges: %v", badgeNames),
		)
	}

	// Add general contextual facts
	for key, fact := range kb.contextualFacts {
		contextParts = append(contextParts, fmt.Sprintf("Fact - %s: %s", key, fact))
	}

	return fmt.Sprintf("%s\n\nProvide a comprehensive and contextually rich response within the application's scope.",
		strings.Join(contextParts, "\n"))
}

// CultourAIClient manages comprehensive AI interactions
type CultourAIClient struct {
	client         *genai.Client
	config         *configs.Config
	rateLimiter    *rate.Limiter
	sessionManager *SessionManager
	knowledgeBase  *ComprehensiveKnowledgeBase
	safetyFilter   *SafetyFilter
}

// SessionManager handles in-memory chat sessions
type SessionManager struct {
	sessions    map[string]*ChatSession
	mutex       sync.RWMutex
	maxSessions int
	sessionTTL  time.Duration
}

// ChatSession represents a single user's chat context
type ChatSession struct {
	ID           string
	UserID       string
	EventID      *string
	Messages     []ChatMessage
	CreatedAt    time.Time
	LastActivity time.Time
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	Role    string
	Content string
	Time    time.Time
}

// NewCultourAIClient creates a new AI client
func NewCultourAIClient(config *configs.Config) (*CultourAIClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GeminiAI.ApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &CultourAIClient{
		client: client,
		config: config,
		rateLimiter: rate.NewLimiter(
			rate.Every(time.Minute), // 100 requests per minute
			100,
		),
		sessionManager: &SessionManager{
			sessions:    make(map[string]*ChatSession),
			maxSessions: 1000,
			sessionTTL:  30 * time.Minute,
		},
		knowledgeBase: NewComprehensiveKnowledgeBase(),
		safetyFilter: &SafetyFilter{
			MinResponseQuality: 0.7,
			MaxResponseLength:  2000,
		},
	}, nil
}

// CreateSession starts a new chat session
func (c *CultourAIClient) CreateSession(userID string, eventID *string) (*ChatSession, error) {
	c.sessionManager.mutex.Lock()
	defer c.sessionManager.mutex.Unlock()

	// Check session limit
	if len(c.sessionManager.sessions) >= c.sessionManager.maxSessions {
		return nil, fmt.Errorf("maximum sessions reached")
	}

	sessionID := generateUniqueSessionID()
	session := &ChatSession{
		ID:           sessionID,
		UserID:       userID,
		EventID:      eventID,
		Messages:     []ChatMessage{},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	c.sessionManager.sessions[sessionID] = session

	return session, nil
}

// GetSession retrieves an existing session
func (c *CultourAIClient) GetSession(sessionID string) (*ChatSession, error) {
	c.sessionManager.mutex.RLock()
	defer c.sessionManager.mutex.RUnlock()

	session, exists := c.sessionManager.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check session expiry
	if time.Since(session.LastActivity) > c.sessionManager.sessionTTL {
		delete(c.sessionManager.sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// AddMessage adds a message to a session
func (c *CultourAIClient) AddMessage(sessionID, role, content string) error {
	c.sessionManager.mutex.Lock()
	defer c.sessionManager.mutex.Unlock()

	session, exists := c.sessionManager.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	message := ChatMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.LastActivity = time.Now()

	return nil
}

// buildContext prepares a comprehensive context for AI interaction
func (c *CultourAIClient) buildContext(session *ChatSession) context.Context {
	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// If session has an event, we can use that to further refine the context
	if session.EventID != nil {
		// Optionally retrieve event details from knowledge base
		event := c.knowledgeBase.GetEvent(fmt.Sprintf("%d", *session.EventID))
		if event != nil {
			// You could add event-specific context preparation here
		}
	}

	// Add cancellation to ensure resources are cleaned up
	go func() {
		// Cleanup after context is done
		<-ctx.Done()
		cancel()
	}()

	return ctx
}

// GenerateResponse generates an AI response
func (c *CultourAIClient) GenerateResponse(sessionID, query string) (string, error) {
	// Rate limit check
	if err := c.rateLimiter.Wait(context.Background()); err != nil {
		return "", fmt.Errorf("rate limit exceeded")
	}

	session, err := c.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	// Prepare context
	context := c.buildContext(session)

	// Generate response
	resp, err := c.client.Models.GenerateContent(context, c.config.GeminiAI.AIModel, genai.Text(query), &genai.GenerateContentConfig{
		Temperature: c.config.GeminiAI.Temperature,
		TopP:        c.config.GeminiAI.TopP,
		TopK:        c.config.GeminiAI.TopK,
		SystemInstruction: &genai.Content{
			Role: "system",
			Parts: []*genai.Part{
				{Text: GetFullSystemPolicy()},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// Validate response

	// Add response to session
	c.AddMessage(sessionID, "assistant", resp.Text())

	return resp.Text(), nil
}

// Helper functions
func generateUniqueSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func (k *ComprehensiveKnowledgeBase) GetEvent(eventID string) *culturalModels.Event {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	return k.events[eventID]
}
