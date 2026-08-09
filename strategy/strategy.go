package strategy

//go:generate go tool stringer -type=RepoExistence -output repoexistence_stringer_generated.go
//go:generate go tool stringer -type=LockExistence -output lockexistence_stringer_generated.go
//go:generate go tool stringer -type=RepoStatus -output repostatus_stringer_generated.go
//go:generate go tool stringer -type=UpdateSpec -output updatespec_stringer_generated.go

type (
	RepoExistence int
	LockExistence int
	RepoStatus    int
	UpdateSpec    int
)

const (
	// REnone means that the repo directory is not existing.
	REnone RepoExistence = iota
	// REexist means that the repo diretocy is existing.
	REexist
)

const (
	// LEnone means that the lock file is not existing.
	LEnone LockExistence = iota
	// LEexist means that the lock file is existing.
	LEexist
)

const (
	// RSunknown means that the repo status is unknown.
	RSunknown RepoStatus = iota
	// RSconflict means that the repo status and the content of the lock file do not match.
	RSconflict
	// RSmatch means that the repo status and the content of the lock file do match.
	RSmatch
)

const (
	// USunspec means that no update strategy is specified.
	USunspec UpdateSpec = iota
	// USforce means that update strategy is the forced updates.
	USforce
	// USretry means that continues processing even without repository updates.
	USretry
	// USnoupdate means that continues processing without repository updates.
	USnoupdate
	// USuninstall means that executes uninstall.
	USuninstall
	// USremove means that executes uninstall and remove the repository.
	USremove
)

//go:generate go tool stringer -type=Type -output type_stringer_generated.go

type Type int

const (
	Tunknown Type = iota
	// TinitFromEmpty clones the repo and create a new lock.
	TinitFromEmpty
	// TinitFromEmptyToLock clones the repo and checkout.
	TinitFromEmptyToLock
	// TinitFromEmptyToLatest clones the repo and update lock.
	TinitFromEmptyToLatest
	// TcreateLock creates a lock and reflects the current status to the lock.
	TcreateLock
	// TcreateLatestLock creates a lock, pulls latest and reflects the status to the lock.
	TcreateLatestLock
	// TupdateToLock checkout to the lock.
	TupdateToLock
	// TupdateToLatestWithLock checkout to the latest and reflects the status to the lock.
	TupdateToLatestWithLock
	// Tnoop does nothing.
	Tnoop
	// Tretry does nothing but continues processing.
	Tretry
	// Tnoupdate does nothing but continues processing.
	Tnoupdate
	// Tremove removes the repo.
	Tremove
)

func NewFact(re RepoExistence, le LockExistence, rs RepoStatus, us UpdateSpec) Fact {
	return Fact{
		RExist:  re,
		LExist:  le,
		RStatus: rs,
		USpec:   us,
	}
}

type Fact struct {
	RExist  RepoExistence
	LExist  LockExistence
	RStatus RepoStatus
	USpec   UpdateSpec
}

// factKey is the composite key for the strategy lookup table.
type factKey struct {
	re RepoExistence
	le LockExistence
	rs RepoStatus
	us UpdateSpec
}

// strategyTable maps all known (RepoExistence, LockExistence, RepoStatus, UpdateSpec)
// combinations to the appropriate strategy Type.
// Entries with USuninstall, USremove, or USnoupdate are handled by short-circuit
// rules in SelectStrategy before this table is consulted.
var strategyTable = map[factKey]Type{
	// REnone + LEnone: no repo, no lock — always init from empty regardless of status or spec
	{REnone, LEnone, RSunknown, USunspec}: TinitFromEmpty,
	{REnone, LEnone, RSunknown, USforce}:  TinitFromEmpty,
	{REnone, LEnone, RSunknown, USretry}:  TinitFromEmpty,
	{REnone, LEnone, RSconflict, USunspec}: TinitFromEmpty,
	{REnone, LEnone, RSconflict, USforce}:  TinitFromEmpty,
	{REnone, LEnone, RSconflict, USretry}:  TinitFromEmpty,
	{REnone, LEnone, RSmatch, USunspec}: TinitFromEmpty,
	{REnone, LEnone, RSmatch, USforce}:  TinitFromEmpty,
	{REnone, LEnone, RSmatch, USretry}:  TinitFromEmpty,

	// REnone + LEexist: no repo but lock exists
	{REnone, LEexist, RSunknown, USunspec}: TinitFromEmptyToLock,
	{REnone, LEexist, RSunknown, USretry}:  TinitFromEmptyToLock,
	{REnone, LEexist, RSunknown, USforce}:  TinitFromEmptyToLatest,
	{REnone, LEexist, RSconflict, USunspec}: TinitFromEmptyToLock,
	{REnone, LEexist, RSconflict, USretry}:  TinitFromEmptyToLock,
	{REnone, LEexist, RSconflict, USforce}:  TinitFromEmptyToLatest,
	{REnone, LEexist, RSmatch, USunspec}: TinitFromEmptyToLock,
	{REnone, LEexist, RSmatch, USretry}:  TinitFromEmptyToLock,
	{REnone, LEexist, RSmatch, USforce}:  TinitFromEmptyToLatest,

	// REexist + LEnone: repo exists but no lock
	{REexist, LEnone, RSunknown, USunspec}: TcreateLock,
	{REexist, LEnone, RSunknown, USretry}:  TcreateLock,
	{REexist, LEnone, RSunknown, USforce}:  TcreateLatestLock,
	{REexist, LEnone, RSconflict, USunspec}: TcreateLock,
	{REexist, LEnone, RSconflict, USretry}:  TcreateLock,
	{REexist, LEnone, RSconflict, USforce}:  TcreateLatestLock,
	{REexist, LEnone, RSmatch, USunspec}: TcreateLock,
	{REexist, LEnone, RSmatch, USretry}:  TcreateLock,
	{REexist, LEnone, RSmatch, USforce}:  TcreateLatestLock,

	// REexist + LEexist + RSconflict: repo and lock exist but differ
	{REexist, LEexist, RSconflict, USunspec}: TupdateToLock,
	{REexist, LEexist, RSconflict, USretry}:  TupdateToLock,
	{REexist, LEexist, RSconflict, USforce}:  TupdateToLatestWithLock,

	// REexist + LEexist + RSmatch: repo and lock exist and match
	{REexist, LEexist, RSmatch, USunspec}: Tnoop,
	{REexist, LEexist, RSmatch, USretry}:  Tretry,
	{REexist, LEexist, RSmatch, USforce}:  TupdateToLatestWithLock,

	// REexist + LEexist + RSunknown: repo and lock exist but status unknown
	{REexist, LEexist, RSunknown, USforce}: TupdateToLatestWithLock,
}

// SelectStrategy determines the update strategy based on the current fact.
//
// Short-circuit rules (evaluated before the lookup table):
//   - USnoupdate → Tnoupdate
//   - USremove   → Tremove
//   - USuninstall → Tnoop
func (f Fact) SelectStrategy() Type {
	switch f.USpec {
	case USnoupdate:
		return Tnoupdate
	case USremove:
		return Tremove
	case USuninstall:
		return Tnoop
	}

	if t, ok := strategyTable[factKey{f.RExist, f.LExist, f.RStatus, f.USpec}]; ok {
		return t
	}
	return Tunknown
}

func (t Type) Runner(c RunnerConfig) Runner {
	switch t {
	case TinitFromEmpty:
		return NewInitFromEmptyRunner(c)
	case TinitFromEmptyToLock:
		return NewInitFromEmptyToLockRunner(c)
	case TinitFromEmptyToLatest:
		return NewInitFromEmptyToLatestRunner(c)
	case TcreateLock:
		return NewCreateLockRunner(c)
	case TcreateLatestLock:
		return NewCreateLatestLockRunner(c)
	case TupdateToLock:
		return NewUpdateToLockRunner(c)
	case TupdateToLatestWithLock:
		return NewUpdateToLatestWithLock(c)
	case Tnoop:
		return NewNoopRunner()
	case Tretry:
		return NewRetryRunner()
	case Tnoupdate:
		return NewNoUpdateRunner()
	case Tremove:
		return NewRemoveRunner(c)
	default:
		return NewUnknownRunner()
	}
}
