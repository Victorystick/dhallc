let Nats1 = [ 1, 2 ] # [ 3, 4 ]

let Ints = [ +1, -2 ] # [ -3, +4 ]

let Texts = [ "I'm" ++ Natural/show 1 ++ "!" ]

let Empty = [] : List Text

let NatLen = List/length Natural Nats1

let Nats = Nats1 # [ Natural/subtract NatLen 1, Natural/subtract 1 NatLen ]

in  { Nats, NatLen, Ints, Texts, Empty }
