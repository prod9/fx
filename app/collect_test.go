package app

import (
	"embed"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.sql
var testMigrations embed.FS

func TestCollectCommands_UnionsTree(t *testing.T) {
	var (
		parentCmd = &cobra.Command{Use: "parent"}
		childCmd  = &cobra.Command{Use: "child"}

		child = Build().Name("child").Command(childCmd)
		root  = Build().Name("root").Command(parentCmd).Mount(child)
	)

	cmds := CollectCommands(root.App())
	require.ElementsMatch(t, []*cobra.Command{parentCmd, childCmd}, cmds)
}

func TestCollectJobs_UnionsTree(t *testing.T) {
	var (
		parentJob = &worker.TestJob{Arg: "parent"}
		childJob  = &worker.TestJob{Arg: "child"}

		child = Build().Name("child").Job(childJob)
		root  = Build().Name("root").Job(parentJob).Mount(child)
	)

	jobs := CollectJobs(root.App())
	require.ElementsMatch(t, []worker.Interface{parentJob, childJob}, jobs)
}

func TestCollectFragment_PreservesChildControllers(t *testing.T) {
	var (
		child = Build().Name("child").Controllers(controllers.Home{})
		root  = Build().Name("root").Mount(child)
	)

	frag := CollectFragment(root.App())
	require.False(t, frag.IsEmpty(),
		"a tree whose controllers live in a mounted child is not empty")
}

func TestCollectFragment_EmptyTree(t *testing.T) {
	root := Build().Name("root").Mount(Build().Name("child"))

	frag := CollectFragment(root.App())
	require.True(t, frag.IsEmpty())
}

func TestRegisterMigrations_FeedsLoadAuto(t *testing.T) {
	t.Chdir(t.TempDir())

	var (
		child = Build().Name("child").EmbedMigrations(testMigrations)
		root  = Build().Name("root").Mount(child)
	)

	RegisterMigrations(root.App())

	migs, err := migrator.LoadAuto(config.Configure())
	require.NoError(t, err)

	names := make([]string, 0, len(migs))
	for _, mig := range migs {
		names = append(names, mig.Name)
	}
	require.Contains(t, names, "202601010101_collect_fixture")
}
