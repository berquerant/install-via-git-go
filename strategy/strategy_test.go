package strategy_test

import (
	"berquerant/install-via-git-go/strategy"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFact(t *testing.T) {
	t.Run("SelectStrategy", func(t *testing.T) {
		allRepoExistence := []strategy.RepoExistence{
			strategy.REnone,
			strategy.REexist,
		}
		allLockExistence := []strategy.LockExistence{
			strategy.LEnone,
			strategy.LEexist,
		}
		allRepoStatus := []strategy.RepoStatus{
			strategy.RSunknown,
			strategy.RSconflict,
			strategy.RSmatch,
		}
		// allUpdateSpec := []strategy.UpdateSpec{
		// 	strategy.USunspec,
		// 	strategy.USforce,
		// 	strategy.USretry,
		// 	strategy.USnoupdate,
		// }

		testFact := func(
			want strategy.Type,
			res []strategy.RepoExistence,
			les []strategy.LockExistence,
			rss []strategy.RepoStatus,
			uss []strategy.UpdateSpec,
		) func(*testing.T) {
			return func(t *testing.T) {
				for _, re := range res {
					for _, le := range les {
						for _, rs := range rss {
							for _, us := range uss {
								title := fmt.Sprintf("%s_%s_%s_%s", re, le, rs, us)
								fact := strategy.NewFact(re, le, rs, us)
								want := want
								t.Run(title, func(t *testing.T) {
									got := fact.SelectStrategy()
									assert.Equal(t, want, got, "want %s got %s", want, got)
								})
							}
						}
					}
				}
			}
		}

		t.Run(
			"retry",
			testFact(
				strategy.Tretry,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEexist},
				[]strategy.RepoStatus{strategy.RSmatch},
				[]strategy.UpdateSpec{strategy.USretry},
			),
		)

		t.Run(
			"noop",
			testFact(
				strategy.Tnoop,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEexist},
				[]strategy.RepoStatus{strategy.RSmatch},
				[]strategy.UpdateSpec{strategy.USunspec},
			),
		)

		t.Run(
			"updateToLatestWithLock",
			testFact(
				strategy.TupdateToLatestWithLock,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEexist},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USforce},
			),
		)

		t.Run(
			"updateToLock",
			testFact(
				strategy.TupdateToLock,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEexist},
				[]strategy.RepoStatus{strategy.RSconflict},
				[]strategy.UpdateSpec{strategy.USunspec},
			),
		)

		t.Run(
			"createLatestLock",
			testFact(
				strategy.TcreateLatestLock,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEnone},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USforce},
			),
		)

		t.Run(
			"createLock",
			testFact(
				strategy.TcreateLock,
				[]strategy.RepoExistence{strategy.REexist},
				[]strategy.LockExistence{strategy.LEnone},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USunspec},
			),
		)

		t.Run(
			"initFromEmptyToLatest",
			testFact(
				strategy.TinitFromEmptyToLatest,
				[]strategy.RepoExistence{strategy.REnone},
				[]strategy.LockExistence{strategy.LEexist},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USforce},
			),
		)

		t.Run(
			"initFromEmptyToLock",
			testFact(
				strategy.TinitFromEmptyToLock,
				[]strategy.RepoExistence{strategy.REnone},
				[]strategy.LockExistence{strategy.LEexist},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USunspec},
			),
		)

		t.Run(
			"initFromEmpty",
			testFact(
				strategy.TinitFromEmpty,
				[]strategy.RepoExistence{strategy.REnone},
				[]strategy.LockExistence{strategy.LEnone},
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USunspec, strategy.USretry, strategy.USforce},
			),
		)

		t.Run(
			"noupdate",
			testFact(
				strategy.Tnoupdate,
				allRepoExistence,
				allLockExistence,
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USnoupdate},
			),
		)

		t.Run(
			"uninstall",
			testFact(
				strategy.Tnoop,
				allRepoExistence,
				allLockExistence,
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USuninstall},
			),
		)

		t.Run(
			"remove",
			testFact(
				strategy.Tremove,
				allRepoExistence,
				allLockExistence,
				allRepoStatus,
				[]strategy.UpdateSpec{strategy.USremove},
			),
		)

	})
}
