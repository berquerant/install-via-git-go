// Code generated by "stringer -type=RepoStatus -output repostatus_stringer_generated.go"; DO NOT EDIT.

package strategy

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RSunknown-0]
	_ = x[RSconflict-1]
	_ = x[RSmatch-2]
}

const _RepoStatus_name = "RSunknownRSconflictRSmatch"

var _RepoStatus_index = [...]uint8{0, 9, 19, 26}

func (i RepoStatus) String() string {
	if i < 0 || i >= RepoStatus(len(_RepoStatus_index)-1) {
		return "RepoStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _RepoStatus_name[_RepoStatus_index[i]:_RepoStatus_index[i+1]]
}
