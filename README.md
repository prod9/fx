# PRODIGY9 FRAMEWORK

A minimalistic Go API application framework.

TODO:
* Tests
* Docs

### Development

To use this in your project, simply do a [git subtree add][0]:

```sh
git subtree add --prefix=fx https://github.com/prod9/fx main
```

* You may edit any files in `fx` as you see fit. (because subtree)
* Do a `git subtree pull` when you want to update `fx` in your repo.
* If you manage to fix something:
  * Isolate the `fx/` changes into its own commit
  * Prefix commit message with `fx:`
  * Use `git subtree push` to push changes back.

### Vanity Server

The `main.go` file in the topmost folder runs a small go application that serves up Go
vanity server so that the import path URL actually works.

A docker image can be build using the `build.py` [Dagger][1] script.

```sh
pip install --upgrade dagger-io==0.6.2 anyio  # first clone
./build.py
```


[0]: https://www.atlassian.com/git/tutorials/git-subtree
[1]: https://dagger.io
