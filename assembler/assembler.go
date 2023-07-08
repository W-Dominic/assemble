package assembler

import(
  "io/ioutil"
  "os/exec"
  "fmt"
)

func Assemble(filePath *string) (string, error) {
  // Compile the file to assembly
  outFile := "out.s"  
  cmd := exec.Command("gcc", "-S", "-masm=intel", "-o", outFile, *filePath)
  err := cmd.Run()
  if err != nil {
    return "", fmt.Errorf("Failed compile %v\n", err)
  }

  // Read and print the assembly code
  assembly, err := ioutil.ReadFile(outFile)
  if err != nil {
    return "", fmt.Errorf("Failed to read aseembly file %v\n", err)
  }

  return string(assembly[:]), nil
}

