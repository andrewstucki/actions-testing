{
  "nodes": {
    "devshell": {
      "inputs": {
        "nixpkgs": [
          "nixpkgs"
        ]
      },
      "locked": {
        "lastModified": 1735644329,
        "narHash": "sha256-tO3HrHriyLvipc4xr+Ewtdlo7wM1OjXNjlWRgmM7peY=",
        "owner": "numtide",
        "repo": "devshell",
        "rev": "f7795ede5b02664b57035b3b757876703e2c3eac",
        "type": "github"
      },
      "original": {
        "owner": "numtide",
        "repo": "devshell",
        "type": "github"
      }
    },
    "flake-parts": {
      "inputs": {
        "nixpkgs-lib": "nixpkgs-lib"
      },
      "locked": {
        "lastModified": 1740872218,
        "narHash": "sha256-ZaMw0pdoUKigLpv9HiNDH2Pjnosg7NBYMJlHTIsHEUo=",
        "owner": "hercules-ci",
        "repo": "flake-parts",
        "rev": "3876f6b87db82f33775b1ef5ea343986105db764",
        "type": "github"
      },
      "original": {
        "owner": "hercules-ci",
        "repo": "flake-parts",
        "type": "github"
      }
    },
    "nixpkgs": {
      "locked": {
        "lastModified": 1741173522,
        "narHash": "sha256-k7VSqvv0r1r53nUI/IfPHCppkUAddeXn843YlAC5DR0=",
        "owner": "NixOS",
        "repo": "nixpkgs",
        "rev": "d69ab0d71b22fa1ce3dbeff666e6deb4917db049",
        "type": "github"
      },
      "original": {
        "id": "nixpkgs",
        "ref": "nixos-unstable",
        "type": "indirect"
      }
    },
    "nixpkgs-lib": {
      "locked": {
        "lastModified": 1740872140,
        "narHash": "sha256-3wHafybyRfpUCLoE8M+uPVZinImg3xX+Nm6gEfN3G8I=",
        "type": "tarball",
        "url": "https://github.com/NixOS/nixpkgs/archive/6d3702243441165a03f699f64416f635220f4f15.tar.gz"
      },
      "original": {
        "type": "tarball",
        "url": "https://github.com/NixOS/nixpkgs/archive/6d3702243441165a03f699f64416f635220f4f15.tar.gz"
      }
    },
    "root": {
      "inputs": {
        "devshell": "devshell",
        "flake-parts": "flake-parts",
        "nixpkgs": "nixpkgs"
      }
    }
  },
  "root": "root",
  "version": 7
}
