# nav

visual navigator for your terminal

# install

```
go install
```

# alias

Put the following line in your `~/.bashrc`:

```
alias nav='nav && cd "`cat /tmp/nav-path`"'
```

# alternative install

Here's what to do when `go install` doesn't work (e.g. on NixOS).

Build `nav` in the package directory:

```
go build
```

Put it somewhere handy, for example:

```
cp nav $HOME/bin/nav
```

Then make sure it's in your path.
