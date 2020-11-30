{ pkgs ? import <nixpkgs> { } }:
with pkgs;
with stdenv;
mkShell {
  name = "autopatchelf-shell";
  buildInputs = [go];
}