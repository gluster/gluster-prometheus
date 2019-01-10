package glusterutils

import (
	"errors"
	"sync"
	"time"
)

// GCache is a wrapper around 'GInterface' object
type GCache struct {
	gd                GInterface
	ttl               time.Duration
	lock              sync.Mutex
	lastCallValueMap  map[string]interface{}
	lastCallTimeMap   map[string]time.Time
	cacheEnabledFuncs map[string]struct{}
}

// NewGCacheWithTTL method creates a new GCache wrapper instance
func NewGCacheWithTTL(gd GInterface, ttl time.Duration) *GCache {
	var gc = new(GCache)
	gc.gd = gd
	gc.ttl = 1 * time.Minute // default to 1 minute
	gc.SetTTL(ttl)
	gc.lastCallValueMap = make(map[string]interface{})
	gc.lastCallTimeMap = make(map[string]time.Time)
	// by default we are enabling cache only for a few functions
	gc.cacheEnabledFuncs = map[string]struct{}{
		"IsLeader":    {},
		"LocalPeerID": {},
		"VolumeInfo":  {},
	}
	return gc
}

// NewGCache method creates a new GCache wrapper instance,
// with 1 minute default time_to_live
func NewGCache(gd GInterface) *GCache {
	return NewGCacheWithTTL(gd, 1*time.Minute)
}

// TTL method returns the current time_to_live duration
func (gc *GCache) TTL() time.Duration {
	return gc.ttl
}

// SetTTL method sets a new time_to_live
func (gc *GCache) SetTTL(ttl time.Duration) {
	// accepts only if the given ttl is greater than a second
	if ttl > time.Second {
		gc.ttl = ttl
	}
}

// EnableCacheForFuncs method will enable caching
// for the given list of functions.
// If the provided function is not there in the existing list, it will be ignored
func (gc *GCache) EnableCacheForFuncs(fNames []string) {
	for _, fName := range fNames {
		gc.cacheEnabledFuncs[fName] = struct{}{}
	}
}

func (gc *GCache) timeForNewCall(funcName string, origFuncName string) (ret bool) {
	if origFuncName == "" {
		origFuncName = funcName
	}
	ret = true
	// if the caching is not enabled for this function, always return true
	// that means, it is always time for a new call
	if _, ok := gc.cacheEnabledFuncs[origFuncName]; !ok {
		return
	}
	if _, ok := gc.lastCallTimeMap[funcName]; !ok {
		return
	}
	nowT := time.Now()
	// if the last called time is BEFORE 'now - ttl', then call again
	if gc.lastCallTimeMap[funcName].Before(nowT.Add(-gc.ttl)) {
		return
	}
	return false
}

// EnableVolumeProfiling method wraps the GInterface.EnableVolumeProfiling call
func (gc *GCache) EnableVolumeProfiling(vInfo Volume) error {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const origName = "EnableVolumeProfiling"
	// caching the result for each volume
	var localName = origName + "-" + vInfo.ID + "-" + vInfo.Name
	var retVal error
	if gc.timeForNewCall(localName, origName) {
		if retVal = gc.gd.EnableVolumeProfiling(vInfo); retVal != nil {
			return retVal
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	return retVal
}

// HealInfo method wraps the GInterface.HealInfo call
func (gc *GCache) HealInfo(vol string) ([]HealEntry, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	// adding the argument[s] also to the 'localName'
	// as we want to cache the function call with each argument
	// it will be wrong to cache the results for only one volume
	// and show the same result throughout for other volumes
	const origName = "HealInfo"
	var localName = origName + "-" + vol
	var retVal []HealEntry
	var err error
	var ok bool
	if gc.timeForNewCall(localName, origName) {
		if retVal, err = gc.gd.HealInfo(vol); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]HealEntry); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// IsLeader method wraps the GInterface.IsLeader call
func (gc *GCache) IsLeader() (bool, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const localName = "IsLeader"
	var retVal bool
	var err error
	var ok bool
	if gc.timeForNewCall(localName, localName) {
		if retVal, err = gc.gd.IsLeader(); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].(bool); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// LocalPeerID method wraps the GInterface.LocalPeerID call
func (gc *GCache) LocalPeerID() (string, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const localName = "LocalPeerID"
	var retVal string
	var err error
	var ok bool
	if gc.timeForNewCall(localName, localName) {
		if retVal, err = gc.gd.LocalPeerID(); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].(string); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// Peers method wraps the GInterface.Peers call
func (gc *GCache) Peers() ([]Peer, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const localName = "Peers"
	var retVal []Peer
	var err error
	var ok bool
	if gc.timeForNewCall(localName, localName) {
		if retVal, err = gc.gd.Peers(); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]Peer); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// Snapshots method wraps the GInterface.Snapshots call
func (gc *GCache) Snapshots() ([]Snapshot, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const localName = "Snapshots"
	var retVal []Snapshot
	var err error
	var ok bool
	if gc.timeForNewCall(localName, localName) {
		if retVal, err = gc.gd.Snapshots(); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]Snapshot); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// VolumeBrickStatus method wraps the GInterface.VolumeBrickStatus call
func (gc *GCache) VolumeBrickStatus(vol string) ([]BrickStatus, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	// caching the results for each volume
	const origName = "VolumeBrickStatus"
	var localName = origName + "-" + vol
	var retVal []BrickStatus
	var err error
	var ok bool
	if gc.timeForNewCall(localName, origName) {
		if retVal, err = gc.gd.VolumeBrickStatus(vol); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]BrickStatus); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// VolumeInfo method wraps the GInterface.VolumeInfo call
func (gc *GCache) VolumeInfo() ([]Volume, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	const localName = "VolumeInfo"
	var retVal []Volume
	var err error
	var ok bool
	if gc.timeForNewCall(localName, localName) {
		if retVal, err = gc.gd.VolumeInfo(); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]Volume); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}

// VolumeProfileInfo method wraps the GInterface.VolumeProfileInfo call
func (gc *GCache) VolumeProfileInfo(vol string) ([]ProfileInfo, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	// caching the results for each volume
	const origName = "VolumeProfileInfo"
	var localName = origName + "-" + vol
	var retVal []ProfileInfo
	var err error
	var ok bool
	if gc.timeForNewCall(localName, origName) {
		if retVal, err = gc.gd.VolumeProfileInfo(vol); err != nil {
			return retVal, err
		}
		// reset the last called time only on a successful call
		gc.lastCallTimeMap[localName] = time.Now()
		gc.lastCallValueMap[localName] = retVal
	}
	if retVal, ok = gc.lastCallValueMap[localName].([]ProfileInfo); !ok {
		err = errors.New("[CacheError] Unable to convert back to a valid return type")
	}
	return retVal, err
}
