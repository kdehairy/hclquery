module {
  block1 {
    attr11 = "a"
    attr12 = "b"
		jsonAttr = jsondecode("{\"name\":\"datetime\",\"image\":\"datetime-image-path\"}")
  }
  block2 {
    attr11 = "a"
    attr12 = "b"
		jsonAttr = jsondecode("[{\"name\":\"datetime\",\"image\":\"datetime-image-path\"}]")
  }
  block3 {
    attr11 = "a"
    attr12 = "b"
		jsonAttr = jsondecode("{\"name\":\"datetime\",\"image\":\"datetime-image-path\", \"env\":[{\"name\":\"name1\", \"value\": \"val1\"}]}")
  }
  attr01 = "x"
  attr02 = "y"
  attr03 = "z"
  attr_obj = {
    a = upper("a")
    b = "b"
    c = "c"
  }
}
