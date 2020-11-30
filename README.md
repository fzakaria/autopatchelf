# autopatchelf

⚠️ This accompanying repository is a _work in progress_ at the moment.

The goal of this repository is to accompany [nix-ld](https://github.com/Mic92/nix-ld) to help select the correct libraries for non-NixOS binaries.

[![asciicast](https://asciinema.org/a/AcR3pbxVRt7leBuF6ds9e8epQ.svg)](https://asciinema.org/a/AcR3pbxVRt7leBuF6ds9e8epQ)

## Usage

For instance to run it through a non-NixOS Ruby
```bash
❯ which ruby
/usr/bin/ruby

❯ go run autopatchelf.go $(which ruby)
```