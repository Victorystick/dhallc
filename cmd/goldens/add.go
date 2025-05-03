package add

func add(x uint) func(uint) uint {
  return func(y uint) uint {
    return x + y
  }
}

var Sum = add(1)(2)
