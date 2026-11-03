package emulator

// Process assertions protect a native process from suspension. The host owns
// the emulator's lifetime, so they need no OS assertion or signal source.
func (r *Runtime) keepAliveMessage(receiver uint64, class, selector string, meta bool) bool {
	if class != "AMSRBSKeepAlive" {
		return false
	}
	if meta {
		switch selector {
		case "keepAliveWithName:", "keepAliveWithName:style:":
			object := r.New(class, nil)
			properties := r.objects[object].Properties
			properties["hostKeepAlive"] = 1
			properties["name"] = r.arg(2)
			if selector == "keepAliveWithName:style:" {
				properties["style"] = r.arg(3)
			}
			r.set(object)
			return true
		}
		return false
	}
	object := r.objects[receiver]
	if object == nil || object.Properties["hostKeepAlive"] == 0 {
		return false
	}
	switch selector {
	case "name", "style":
		r.set(object.Properties[selector])
	case "invalidate":
		object.Properties["invalidated"] = 1
		r.set(0)
	case ".cxx_destruct":
		r.set(0)
	default:
		return false
	}
	return true
}
