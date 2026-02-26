{ pkgs, ... }: {
  languages.go.enable = true;
  languages.go.package = pkgs.go; # Use default go which should be stable

  packages = [
    pkgs.gh
    pkgs.gopls
    pkgs.delve
    pkgs.act
  ];
}
