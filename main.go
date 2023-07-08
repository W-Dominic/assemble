package main

import (
  "flag"
  "fmt"
  "io/ioutil"
  "os"
  "os/exec"
  "time"
)

func main() {
  filePath := flag.String("file", "./", "Path to file")
  flag.Parse() 
  if *filePath == "./" {
    fmt.Printf("Cannot open file \n")
    os.Exit(1)
  }

  var oldSize int64 = 0

  for {
    fileInfo, err := os.Stat(*filePath)
    if err != nil {
      fmt.Println(err)
      time.Sleep(1 * time.Second)
      continue
    }

    newSize := fileInfo.Size()
    if newSize != oldSize {
      oldSize = newSize

      // Compile the file to assembly
      outFile := "out.s"  
      cmd := exec.Command("gcc", "-S", "-masm=intel", "-o", outFile, *filePath)
      err := cmd.Run()
      if err != nil {
        fmt.Printf("Failed to compile: %s\n", err)
        continue
      }

      // Read and print the assembly code
      assembly, err := ioutil.ReadFile(outFile)
      if err != nil {
        fmt.Printf("Failed to read assembly file: %s\n", err)
        continue
      }
      fmt.Printf("Assembly:\n%s\n", assembly)
    }

    time.Sleep(1 * time.Second)
  }
}
