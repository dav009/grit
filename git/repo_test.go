	"context"
	"os"
	"sync"
var (
	nocleanup  = flag.Bool("nocleanup", false, "don't clean up git state after tests are run")
	shelltrace = flag.Bool("shelltrace", false, "trace shell execution")
)
	patch, err := repo.Patch(c.Digest, "")



	patch, err := src.Patch(commits[0].Digest, "")
}

func TestLFS(t *testing.T) {
	_, err := exec.LookPath("lfs-test-server")
	if err != nil {
		t.Skip("lfs-test-server not installed")
	}
	dir, cleanup := testutil.TempDir(t, "", "")
	if *nocleanup {
		log.Println("directory", dir)
	} else {
		defer cleanup()
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	defer cancel()
	go func() {
		cmd := exec.CommandContext(ctx, "lfs-test-server")
		cmd.Env = []string{
			"LFS_ADMINUSER=user",
			"LFS_ADMINPASS=pass",
			"LFS_CONTENTPATH=" + dir,
		}
		err := cmd.Run()
		if err != nil && err != context.Canceled && !strings.HasSuffix(err.Error(), "signal: killed") {
			log.Panicf("lfs-test-server: %v", err)
		}
		wg.Done()
	}()

	shell(t, dir, `
		mkdir repos

		git init --bare repos/src
		git clone repos/src src
		cd src
		git config user.email you@example.com
		git config user.name "your name"
		git lfs install
		git config -f .lfsconfig lfs.url http://user:pass@localhost:8080
		git add .lfsconfig
		git commit -a -m "lfsconfig"

		echo bigfile >bigfile
		git lfs track bigfile
		git add .
		git commit -a -m "big file"
		git push

		cd ..
		# Create the destination repository. Note that we don't install
		# LFS hooks and instead maintain the pointers manually.
		git init --bare repos/dst
		git clone repos/dst dst
		cd dst
		git config user.email you@example.com
		git config user.name "your name"
		# Manually install the pointer for 'bigfile' into this repository.
		git -C ../src show HEAD:bigfile > bigfile
		git add bigfile
		git commit -m'first commit'
		git push
	`)
	src, err := Open(filepath.Join(dir, "repos/src"), "", "master")
	if err != nil {
		t.Fatal(err)
	}
	ptrs, err := src.ListLFSPointers()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(ptrs), 1; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := ptrs[0], "bigfile"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	dst, err := Open(filepath.Join(dir, "repos/dst"), "", "master")
	if err != nil {
		t.Fatal(err)
	}
	if err := dst.CopyLFSObject(src, ptrs[0]); err != nil {
		t.Fatal(err)
	}
	if *shelltrace {
		cmd.Stderr = os.Stderr
	}
		if *shelltrace {
			t.Fatal("script failed")
		}