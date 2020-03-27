package schema


// Definition for a generator
HofGen :: {
  // Base directory for the output
  Outdir: string | *"./"

  // "Global" input, merged with out replacing onto the files
  In: { ... } | * {...}

  // The list fo files for hof to generate
  Out: [...GenFile] | *[...]

  // Subgenerators for composition
  Generators: [...HofGen] | *{...}

  //
  // For file based generators
  //

  // Used for indexing into the vendor directory...
  PackageName: string | * ""

  // Base directory of entrypoint templates to load
  TemplatesDir: string | * ""

  // Base directory of partial templatess to load
  PartialsDir: string | * ""

  // Filepath globs for static files to load
  StaticGlobs: [...string] | * ""

  // Open for whatever else you may need
  //   often hidden fields are used
  ...
} 

// A file which should be generated by hof
GenFile :: {
  // The local input data
  In: { ... }

  // The full path to the output location
  Filename: string
  // for starting the migration
  Filepath: string
  Filepath: Filename

  //
  // Template parameters
  //

  // The template contents
  Template: string | *""

  // Relative name from TemplatesDir
  TemplateFile: string | *""

  // System params
  TemplateSystem: *"text/template" | "mustache"

  //
  // Template delimiters
  //

  // The default delimiters
  // You should change these when using alternative style like jinjas {% ... %} 
  // They also need to be different when using the swap system
  LHS_D: string | *"{{"
  RHS_D: string | *"}}"
  LHS2_D: LHS_D
  RHS2_D: RHS_D
  LHS3_D: string | *"{{{"
  RHS3_D: string | *"}}}"

  // These are the same as the default becuase
  // the current template systems require these.
  //   So these should really never change or be overriden until there is a new template system
  //     supporting setting the delimiters dynamicalldelimiters dynamicallyy
  LHS_S: string | *"{{"
  RHS_S: string | *"}}"
  LHS2_S: LHS_D
  RHS2_S: RHS_D
  LHS3_S: string | *"{{{"
  RHS3_S: string | *"}}}"

  // The temporary delims to replace swap with while also swapping
  // the defaults you set to the swap that is required by the current templet systems
  LHS_S: string | *"{{"
  RHS_S: string | *"}}"
  LHS2_S: LHS_D
  RHS2_S: RHS_D
  LHS3_S: string | *"{{{"
  RHS3_S: string | *"}}}"

  // Swap delimiters, becuase the template systems use `{{` and `}}`
  //   and if you want to preserve those, we need three sets of delimiters
  Alt: bool | *false
  // for starting the migration
  SwapDelims: bool
  SwapDelims: Alt

  // Open for whatever else you may need, not sure about this case though
  ...
}