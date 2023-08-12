terraform {
  required_providers {
    teraswitch = {
      source = "teraswitch/teraswitch"
    }
  }
}

provider "teraswitch" {
  # https://beta.tsw.io/ => Settings => Developer
  api_token = "... YOUR TOKEN HERE ..."
}
