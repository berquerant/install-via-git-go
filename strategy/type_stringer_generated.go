// Code generated by "stringer -type=Type -output type_stringer_generated.go"; DO NOT EDIT.

package strategy

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Tunknown-0]
	_ = x[TinitFromEmpty-1]
	_ = x[TinitFromEmptyToLock-2]
	_ = x[TinitFromEmptyToLatest-3]
	_ = x[TcreateLock-4]
	_ = x[TcreateLatestLock-5]
	_ = x[TupdateToLock-6]
	_ = x[TupdateToLatestWithLock-7]
	_ = x[Tnoop-8]
	_ = x[Tretry-9]
}

const _Type_name = "TunknownTinitFromEmptyTinitFromEmptyToLockTinitFromEmptyToLatestTcreateLockTcreateLatestLockTupdateToLockTupdateToLatestWithLockTnoopTretry"

var _Type_index = [...]uint8{0, 8, 22, 42, 64, 75, 92, 105, 128, 133, 139}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
