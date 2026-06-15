package core

import (
	"errors"
	"time"
)

// ErrNoActiveSession is returned when an operation targets an interactive
// session that does not exist in the Engine's active state map.
var ErrNoActiveSession = errors.New("no active interactive session for the given key")

// SessionDetail provides a public snapshot of an active interactive session,
// suitable for display in the desktop client's Agent Management page.
type SessionDetail struct {
	SessionKey    string    // e.g. "feishu:chat123:user456"
	AgentSessionID string   // agent's session ID (may be empty if not yet established)
	AgentType     string    // "claudecode", "codex", etc.
	ProjectName   string    // engine name (= project name)
	Platform      string    // "feishu", "telegram", etc.
	ChatName      string    // display name from UserMeta
	UserName      string    // display name from UserMeta
	Status        string    // "active", "idle", "error"
	CreatedAt     time.Time // session creation time (if known)
	UpdatedAt     time.Time // session last update time (if known)
	QueueDepth    int       // number of pending messages
}

// QueuedMessageInfo provides a public snapshot of a message queued in an
// interactive session, suitable for display in the task queue view.
type QueuedMessageInfo struct {
	Platform string    // source platform name
	UserName string    // who sent the message
	Content  string    // truncated content (max 60 chars)
	QueuedAt time.Time // approximate time when queued (best-effort)
}

// ListActiveInteractiveStates returns details for all currently active
// interactive sessions. Each entry corresponds to one entry in the
// Engine.interactiveStates map.
func (e *Engine) ListActiveInteractiveStates() []SessionDetail {
	e.interactiveMu.Lock()
	statesCopy := make(map[string]*interactiveState, len(e.interactiveStates))
	for k, v := range e.interactiveStates {
		statesCopy[k] = v
	}
	e.interactiveMu.Unlock()

	details := make([]SessionDetail, 0, len(statesCopy))
	for key, state := range statesCopy {
		detail := SessionDetail{
			SessionKey:  key,
			ProjectName: e.ProjectName(),
			Status:      "active",
		}

		state.mu.Lock()
		if state.agentSession != nil {
			detail.AgentSessionID = state.agentSession.CurrentSessionID()
			detail.QueueDepth = len(state.pendingMessages)
		}
		detail.Platform = state.platform.Name()
		if state.agent != nil {
			detail.AgentType = state.agent.Name()
		}
		state.mu.Unlock()

		// Try to find a corresponding Session for timestamps
		sessionKey := key
		sessions := e.sessions.AllSessions()
		idToKey, _ := e.sessions.SessionKeyMap()
		for _, s := range sessions {
			if idToKey[s.ID] == sessionKey {
				detail.CreatedAt = s.CreatedAt
				detail.UpdatedAt = s.GetUpdatedAt()
				break
			}
		}

		// Get user meta for display names
		meta := e.sessions.GetUserMeta(key)
		if meta != nil {
			detail.ChatName = meta.ChatName
			detail.UserName = meta.UserName
		}

		details = append(details, detail)
	}

	return details
}

// GetQueuedMessages returns the queued messages for the interactive session
// identified by sessionKey. Returns nil if the session is not active.
func (e *Engine) GetQueuedMessages(sessionKey string) []QueuedMessageInfo {
	e.interactiveMu.Lock()
	state, ok := e.interactiveStates[sessionKey]
	e.interactiveMu.Unlock()

	if !ok || state == nil {
		return nil
	}

	state.mu.Lock()
	msgs := state.pendingMessages
	state.mu.Unlock()

	result := make([]QueuedMessageInfo, 0, len(msgs))
	for _, m := range msgs {
		content := m.content
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		result = append(result, QueuedMessageInfo{
			Platform: m.platform.Name(),
			UserName: m.userName,
			Content:  content,
			// queuedMessage doesn't carry a timestamp; use zero as placeholder
			QueuedAt: time.Time{},
		})
	}

	return result
}

// StopInteractiveSession stops the interactive session identified by sessionKey.
// It calls cleanupInteractiveState to close the agent session and remove the
// state entry. Returns an error if no active session exists for that key.
func (e *Engine) StopInteractiveSession(sessionKey string) error {
	e.interactiveMu.Lock()
	_, ok := e.interactiveStates[sessionKey]
	e.interactiveMu.Unlock()

	if !ok {
		return ErrNoActiveSession
	}

	e.cleanupInteractiveState(sessionKey)
	return nil
}
