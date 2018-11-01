ptmux
============

[![Coverage Status](https://coveralls.io/repos/github/pocke/ptmux/badge.svg?branch=master)](https://coveralls.io/github/pocke/ptmux?branch=master)
[![Build Status](https://travis-ci.org/pocke/ptmux.svg?branch=master)](https://travis-ci.org/pocke/ptmux)


Requirements
---

* Go 1.11 or later
* `set -g base-index 1` in your `.tmux.conf`
* tmux 2.0 or newer

Installation
-----------

```sh
export GO111MODULE=on
go get github.com/pocke/ptmux
```

<!-- Or download a binary from [Latest release](https://github.com/pocke/ptmux/releases/latest). -->


Usage
-----------

### Configure

Edit `~/.config/ptmux/PROFILE_NAME.yaml`

```yaml
# Example
root: ~/path/to/your/project/dir
env:
  DATABASE_URL: 'mysql2://username:topsecret@localhost/dbname'
windows:
  - panes:
    - command: 'bin/rails s'
    - command: 'bundle exec sidekiq'
  - panes:
    - command: 'vim'
    - command: 'bundle exec guard'
```


### Command line


```sh
$ ptmux PROFILE_NAME
```

Links
------

- [tmuxinator の代替ツールを作っている話 - pockestrap](http://pocke.hatenablog.com/entry/2016/11/13/233258)


License
-------

These codes are licensed under CC0.

[![CC0](http://i.creativecommons.org/p/zero/1.0/88x31.png "CC0")](http://creativecommons.org/publicdomain/zero/1.0/deed.en)
