package main

import (
  "fmt"
  "testing"
  "os"
)

// TestChaosExporter is a sample test function
func TestChaosExporter(t *testing.T) {
  fmt.Println("..Test Chaos Exporter..")
}

func TestGetOpenebsEnv(t *testing.T) {
  // Set Enviroment variable
  os.Setenv("_UTEST_OPENEBS", "utestopenebs")

  tests := map[string]struct {
    inputkey          string
    fallbackvalue     string
    returnvalue       string
  }{
    "Test Positive-1":{
      inputkey:       "_UTEST_OPENEBS", 
      fallbackvalue:  "N/A",
      returnvalue:    "utestopenebs",
    },
    "Test Negative-1":{
      inputkey:       "_RUTEST_OPENEBS", 
      fallbackvalue:  "rutest_openebs",
      returnvalue:    "rutest_openebs",
    },
  }

  for name, mock := range tests {
    name, mock := name, mock
    t.Run(name, func(t *testing.T){
      actualresult := getOpenebsEnv(mock.inputkey, mock.fallbackvalue)
      if mock.returnvalue != actualresult {
        t.Fatalf("Test %q failed: expected value=%q, actual value=%q ", name, mock.returnvalue, actualresult)
      }
    })
  //Unset the Environment variable used for unit test
  os.Unsetenv("_UTEST_OPENEBS")
  }
}