# OpenAPI Utils
Simple utils to work with OpenAPI files

## oas-yaml-combine

Takes input .yaml file and follows references to inline everything in a single output .yaml file
Tool can resolve local refs (refs pointing to objects in the same file) and remote refs (refs pointing to object in other files)

### building

- to build executables (linux & windows) run from project root `./scripts/build_cmd.sh`
- to build shared library (linux only atm) run from project root `./scripts/build_lib.sh`

### executable arguments

- `help` - help
- `input-file` - path to input file that has refs that need to be resolved. When not provided, stdin is used
- `output-file` - path to output file to which input file with resolved refs should be saved. When not provided, stdout is used
- `ref-directory` - path to directory where files containing remote refs are stored. When not provided a directory of input-file is used
- `inline-local` - (default: `false`) when set to `true` local refs are replaced with local objects, otherwise local refs stay in place
- `inline-remote` - (default: `false`) when set to `true` remote refs are replaced with remote objects, otherwise remote refs stay in place
- `keep-local` - (default: `false`) when set to `true` along with `inline-local` keeps local reference objects after inlining, otherwise deletes them. When set to `true` with `inline-local` set to false does nothing to prevent from making dangling local references, and therefore creating incorrect specifications