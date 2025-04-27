let user = "bill"

in  { Home = "/home/${user}"
    , PrivateKey = "/home/${user}/.ssh/id_ed25519"
    , PublicKey = "/home/${user}/.ssh/id_ed25519.pub"
    }
