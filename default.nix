with import <nixpkgs> {}; with goPackages;

buildGoPackage rec {
  name = "nav";
  buildInputs = [ termbox-go ];
  goPackagePath = "github.com/mkovacs/nav";
}
