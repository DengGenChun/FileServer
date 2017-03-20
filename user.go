package main

import (
	"sync"
	"time"
)

// UsrManager manage all users that access this server.
// Once user access server frequently,
// then the user's access will be denied.
// If user's count of access more than AccessCountLimit in TimeLimitInSecond,
// then the user's access will be denied in TimeBoforeRestoreInSecond.
type UsrManager struct {
	sync.RWMutex

	// The map to store the user's access data
	usrMap map[string]*Usr
}

// Usr struct to record the user address, access count...
type Usr struct {
	sync.RWMutex

	address        string    // The ip address of user
	accessCount    int       // The access count in limit time
	accessCostTime int64     // The time that the total access cost
	lastAccessTime time.Time // The time of user's last access

	// The flags whether user can access server, true by default.
	// The user can access server only when these two flag both are true.
	canAccessAuto bool // controled by server automatically.
	canAccessOpt  bool // controled by administrator optional.
}

// Must be called once UsrManager was created.
func (u *UsrManager) init() {
	// Initialize the map
	u.usrMap = make(map[string]*Usr)

	go u.expiredCheck()
}

// Check and remove user from map every TimeForExpiredCheckInSecond
// when the user no longer access server in TimeForUserToExistInSecond.
func (u *UsrManager) expiredCheck() {
	timer := time.NewTimer(time.Duration(TimeForExpiredCheckInSecond) * time.Second)
	for {
		select {
		case <-timer.C:
			u.Lock()
			// Loop through the map.
			for addr, usr := range u.usrMap {
				// If the user was denied to access server forerver by administrator,
				// don't delete it from map.
				if false == usr.canAccessOpt {
					continue
				}

				// If the user do not access in TimeForUserToExistInSecond.
				timeDiff := time.Now().UnixNano() - usr.lastAccessTime.UnixNano()
				if timeDiff > TimeForUserToExistInSecond*1e9 {
					// Delete user from map.
					delete(u.usrMap, addr)
				}
			}
			u.Unlock()
		}
	}
}

// Check whether the user with usrAddr can access server.
// Return true when it can, s
func (u *UsrManager) canAccess(usrAddr string) bool {
	// Get user from map
	usr := u.getUsr(usrAddr)
	// User attempt to access server
	usr.accessServer()

	return usr.canAccess()
}

// Set access permission for a user.
// The function should be called by administrator.
func (u *UsrManager) setAccess(usrAddr string, isAccess bool) {
	// Get user from map
	usr := u.getUsr(usrAddr)
	// Set access permission for user
	usr.setOptionalAccess(isAccess)
}

// Return a user with given addr from usrMap.
// If the specific user do not exist, create a new one
func (u *UsrManager) getUsr(usrAddr string) *Usr {
	u.RLock()
	usr, ok := u.usrMap[usrAddr] // Get user from map
	u.RUnlock()
	if !ok || usr == nil { // If user do not exist
		u.Lock()
		usr = new(Usr)          // Create a new user
		usr.init(usrAddr)       // Initialize the user
		u.usrMap[usrAddr] = usr // Store user into map
		u.Unlock()
	}
	return usr
}

// Must be called once Usr was created.
func (u *Usr) init(addr string) {
	u.address = addr
	u.accessCount = 0
	u.accessCostTime = 0.0
	u.lastAccessTime = time.Now()
	u.canAccessAuto = true // User can access server by default.
	u.canAccessOpt = true  // User can access server by default.
}

// User attempt to access server.
// Detect whether user is access frequently,
// If it is, set user can not access server any more.
// After a fixed time, the user should be can access server again.
func (u *Usr) accessServer() {
	// Do nothing when the user is already be denied accessing.
	if !u.canAccess() {
		return
	}

	u.Lock()
	// Access count plus 1.
	u.accessCount++
	// The time, the total access cost,
	// plus the time difference  between last access and this access.
	u.accessCostTime += time.Now().UnixNano() - u.lastAccessTime.UnixNano()
	// Update the last access time.
	u.lastAccessTime = time.Now()

	// Check whether user should be denied accessing.
	if u.accessCount > AccessCountLimit {
		// If the user access frequently.
		if u.accessCostTime < TimeLimitInSecond*1e9 {
			// Set the user can not access server any more.
			u.canAccessAuto = false
			// After timeBoforeRestoreInSecon,
			// Set the user can access server again.
			go func() {
				time.AfterFunc(time.Duration(TimeBoforeRestoreInSecond)*time.Second, func() {
					u.Lock()
					u.canAccessAuto = true
					u.Unlock()
				})
			}()
		}
		// Reset the access count and access cost time for next turn.
		u.accessCount = 0
		u.accessCostTime = 0
	}
	u.Unlock()
}

// Return true when the user can access server, or false.
// The user can access server only when these two flag both are true.
func (u *Usr) canAccess() bool {
	u.RLock()
	isOk := u.canAccessAuto && u.canAccessOpt
	u.RUnlock()
	return isOk
}

// This function should be called by administrator.
func (u *Usr) setOptionalAccess(isAccess bool) {
	u.Lock()
	u.canAccessOpt = isAccess
	u.Unlock()
}
