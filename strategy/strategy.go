package strategy

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=RepoExistence -output repoexistence_stringer_generated.go
//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=LockExistence -output lockexistence_stringer_generated.go
//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=RepoStatus -output repostatus_stringer_generated.go
//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=UpdateSpec -output updatespec_stringer_generated.go

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
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=Type -output type_stringer_generated.go

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

func (f Fact) SelectStrategy() Type {
	switch f.RExist {
	case REnone:
		switch f.LExist {
		case LEnone:
			return TinitFromEmpty
		case LEexist:
			switch f.USpec {
			case USunspec, USretry:
				return TinitFromEmptyToLock
			case USforce:
				return TinitFromEmptyToLatest
			}
		}
	case REexist:
		switch f.LExist {
		case LEnone:
			switch f.USpec {
			case USunspec, USretry:
				return TcreateLock
			case USforce:
				return TcreateLatestLock
			}
		case LEexist:
			switch f.RStatus {
			case RSconflict:
				switch f.USpec {
				case USunspec, USretry:
					return TupdateToLock
				case USforce:
					return TupdateToLatestWithLock
				}
			case RSmatch:
				switch f.USpec {
				case USunspec:
					return Tnoop
				case USretry:
					return Tretry
				case USforce:
					return TupdateToLatestWithLock
				}
			default:
				switch f.USpec {
				case USforce:
					return TupdateToLatestWithLock
				}
			}
		}
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
	default:
		return NewUnknownRunner()
	}
}
