# Golang XRandR wrapper [![Go Report Card](https://goreportcard.com/badge/github.com/vcraescu/go-xrandr)](https://goreportcard.com/report/github.com/vcraescu/go-xrandr) [![Build Status](https://travis-ci.com/vcraescu/go-xrandr.svg?branch=master)](https://travis-ci.com/vcraescu/go-xrandr) [![Coverage Status](https://coveralls.io/repos/github/vcraescu/go-xrandr/badge.svg?branch=master)](https://coveralls.io/github/vcraescu/go-xrandr?branch=master)

Golang xrandr wrapper.
The entire functionality is based on xrandr output.

```go
import (
	"github.com/vcraescu/go-xrandr"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	screens, err := xrandr.GetScreens()
	if err != nil {
		panic(err)
	}
	
	spew.Dump(screens)
}
```

will print: 

```
(xrandr.Screens) (len=1 cap=1) {
 (xrandr.Screen) {
  No: (int) 0,
  CurrentResolution: (xrandr.Size) {
   Width: (float32) 8959,
   Height: (float32) 2880
  },
  MinResolution: (xrandr.Size) {
   Width: (float32) 8,
   Height: (float32) 8
  },
  MaxResolution: (xrandr.Size) {
   Width: (float32) 32767,
   Height: (float32) 32767
  },
  Monitors: ([]xrandr.Monitor) (len=3 cap=4) {
   (xrandr.Monitor) {
    ID: (string) (len=6) "HDMI-0",
    Modes: ([]xrandr.Mode) (len=14 cap=16) {
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 3840,
       Height: (float32) 2160
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 30,
        Current: (bool) true,
        Preferred: (bool) true
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 29.97,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 25,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 23.98,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1920,
       Height: (float32) 1200
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1920,
       Height: (float32) 1080
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=7 cap=8) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 50,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 23.98,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.05,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 50.04,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1680,
       Height: (float32) 1050
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1600,
       Height: (float32) 1200
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 1024
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 75.02,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.02,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 800
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.81,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 720
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=3 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 50,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1152,
       Height: (float32) 864
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 75,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1024,
       Height: (float32) 768
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 75.03,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 800,
       Height: (float32) 600
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 75,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.32,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 720,
       Height: (float32) 576
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 50,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 720,
       Height: (float32) 480
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 640,
       Height: (float32) 480
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=3 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 75,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.93,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     }
    },
    Primary: (bool) true,
    Size: (xrandr.Size) {
     Width: (float32) 597,
     Height: (float32) 336
    },
    Connected: (bool) true,
    Resolution: (xrandr.Size) {
     Width: (float32) 5119,
     Height: (float32) 2879
    },
    Position: (xrandr.Position) {
     X: (int) 0,
     Y: (int) 0
    }
   },
   (xrandr.Monitor) {
    ID: (string) (len=7) "eDP-1-1",
    Modes: ([]xrandr.Mode) (len=50 cap=64) {
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 3840,
       Height: (float32) 2160
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) true,
        Preferred: (bool) true
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.98,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 48.02,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 3200,
       Height: (float32) 1800
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 2880,
       Height: (float32) 1620
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 2560,
       Height: (float32) 1600
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 2560,
       Height: (float32) 1440
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 2048,
       Height: (float32) 1536
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1920,
       Height: (float32) 1440
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1856,
       Height: (float32) 1392
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.01,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1792,
       Height: (float32) 1344
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.01,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 2048,
       Height: (float32) 1152
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.98,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.9,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.91,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1920,
       Height: (float32) 1200
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1920,
       Height: (float32) 1080
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.01,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.93,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1600,
       Height: (float32) 1200
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1680,
       Height: (float32) 1050
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1400,
       Height: (float32) 1050
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.98,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1600,
       Height: (float32) 900
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.82,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 1024
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.02,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1400,
       Height: (float32) 900
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 960
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1440,
       Height: (float32) 810
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1368,
       Height: (float32) 768
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.85,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 800
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.97,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.81,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.91,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1280,
       Height: (float32) 720
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.86,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.74,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1024,
       Height: (float32) 768
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.04,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 960,
       Height: (float32) 720
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 928,
       Height: (float32) 696
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.05,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 896,
       Height: (float32) 672
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.01,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 1024,
       Height: (float32) 576
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.9,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.82,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 960,
       Height: (float32) 600
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.93,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 960,
       Height: (float32) 540
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.99,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.63,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.82,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 800,
       Height: (float32) 600
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=3 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.32,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 56.25,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 840,
       Height: (float32) 525
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.01,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 864,
       Height: (float32) 486
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.92,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.57,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 700,
       Height: (float32) 525
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.98,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 800,
       Height: (float32) 450
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.95,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.82,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 640,
       Height: (float32) 512
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.02,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 700,
       Height: (float32) 450
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.96,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 640,
       Height: (float32) 480
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.94,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 720,
       Height: (float32) 405
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.51,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 58.99,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 684,
       Height: (float32) 384
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.85,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 640,
       Height: (float32) 400
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.88,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.98,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 640,
       Height: (float32) 360
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=4 cap=4) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.86,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.83,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.84,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.32,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 512,
       Height: (float32) 384
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 512,
       Height: (float32) 288
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.92,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 480,
       Height: (float32) 270
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.63,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.82,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 400,
       Height: (float32) 300
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.32,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 56.34,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 432,
       Height: (float32) 243
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.92,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.57,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 320,
       Height: (float32) 240
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=1 cap=1) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 60.05,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 360,
       Height: (float32) 202
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.51,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.13,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     },
     (xrandr.Mode) {
      Resolution: (xrandr.Size) {
       Width: (float32) 320,
       Height: (float32) 180
      },
      RefreshRates: ([]xrandr.RefreshRate) (len=2 cap=2) {
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.84,
        Current: (bool) false,
        Preferred: (bool) false
       },
       (xrandr.RefreshRate) {
        Value: (xrandr.RefreshRateValue) 59.32,
        Current: (bool) false,
        Preferred: (bool) false
       }
      }
     }
    },
    Primary: (bool) false,
    Size: (xrandr.Size) {
     Width: (float32) 346,
     Height: (float32) 194
    },
    Connected: (bool) true,
    Resolution: (xrandr.Size) {
     Width: (float32) 3840,
     Height: (float32) 2160
    },
    Position: (xrandr.Position) {
     X: (int) 5119,
     Y: (int) 0
    }
   },
   (xrandr.Monitor) {
    ID: (string) (len=6) "DP-1-1",
    Modes: ([]xrandr.Mode) <nil>,
    Primary: (bool) false,
    Size: (xrandr.Size) {
     Width: (float32) 0,
     Height: (float32) 0
    },
    Connected: (bool) false,
    Resolution: (xrandr.Size) {
     Width: (float32) 0,
     Height: (float32) 0
    },
    Position: (xrandr.Position) {
     X: (int) 0,
     Y: (int) 0
    }
   }
  }
 }
}
```
