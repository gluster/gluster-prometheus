package glusterutils

import "errors"

// HealInfo gets heal info from glusterd2 using rest api
func (g GD2) HealInfo(vol string) ([]HealEntry, error) {
	return nil, errors.New("not implemented for GD2")
}
