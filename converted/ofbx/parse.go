package ofbx

func parseObjects(root *Element, scene *Scene) (bool, error) {
	objs := findChild(root, "Objects")
	if objs == nil {
		return true, nil
	}

	scene.m_root = NewRoot(*scene, root)
	scene.m_object_map[0] = ObjectPair{&root, scene.m_root}

	object := objs.child
	for object != nil {
		if !isLong(object.first_property) {
			return false, errors.New("Invalid")
		}

		id := object.first_property.value.touint64()
		scene.m_object_map[id] = ObjectPair{object, nullptr}
		object = object.sibling
	}

	for _, iter := range scene.m_object_map {
		var obj *Object

		if iter.second.object == scene.m_root {
			continue
		}

		if iter.second.element.id == "Geometry" {
			last_prop := iter.second.element.first_property
			for last_prop.next != nil {
				last_prop = last_prop.next
			} 
			if last_prop != nil && last_prop.value == "Mesh" {
				obj = parseGeometry(*scene, *iter.second.element)
			}
		} else if iter.second.element.id == "Material" {
			obj = parseMaterial(*scene, *iter.second.element)
		} else if iter.second.element.id == "AnimationStack" {
			obj = parse(*scene, *iter.second.element) //AnimationStack
			if !obj.isError() {
				stack := obj.getValue().(*AnimationStackImpl)
				scene.m_animation_stacks.push_back(stack)
			}
		} else if iter.second.element.id == "AnimationLayer" {
			obj = parse(*scene, *iter.second.element) // AnimationLayer
		} else if iter.second.element.id == "AnimationCurve" {
			obj = parseAnimationCurve(*scene, *iter.second.element)
		} else if iter.second.element.id == "AnimationCurveNode" {
			obj = parse(*scene, *iter.second.element) //AnimationCurveNode
		} else if iter.second.element.id == "Deformer" {
			class_prop = iter.second.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue()
				if v == "Cluster" {
					obj = parseCluster(*scene, *iter.second.element)
				} else if v == "Skin" {
					obj = parse(*scene, *iter.second.element)
				}
			}
		} else if iter.second.element.id == "NodeAttribute" {
			obj = parseNodeAttribute(*scene, *iter.second.element)
		} else if iter.second.element.id == "Model" {
			iter.second.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue()
				if v == "Mesh" {
					obj = parseMesh(*scene, *iter.second.element)
					if !obj.isError() {
						mesh = obj.getValue().(*Mesh)
						scene.m_meshes.push_back(mesh)
						obj = mesh
					}
				} else if v == "LimbNode" {
					obj = parseLimbNode(*scene, *iter.second.element)
				} else if v == "Null" || v == "Root" {
					obj = parse(*scene, *iter.second.element)
				} 
			}
		} else if (iter.second.element.id == "Texture") {
			obj = parseTexture(*scene, *iter.second.element)
		}


		if obj.isError() {
			return false, nil // error?
		}

		val := obj.getValue()
		scene.m_object_map[iter.first].object = val 
		if val != nil {
			scene.m_all_objects.push_back(val)
			val.id = iter.first
		}
	}
	for _, con := range scene.m_connections {
		parent := scene.m_object_map[con.to].object
		child := scene.m_object_map[con.from].object
		if child == nil || parent == nil {
			continue
		}

		ctyp := child.getType()

		switch ctyp {
			case NODE_ATTRIBUTE:
				if parent.node_attribute {
					return false, errors.New("Invalid node attribute")
				}
				parent.node_attribute = child.(*NodeAttribute)
			case ANIMATION_CURVE_NODE:
				if parent.isNode() {
					node := child.(*AnimationCurveNode)
					node.bone = parent
					node.bone_link_property = con.property
				}
		}

		switch (parent.getType()) {
			case MESH: {
				mesh := parent.(*MeshImpl)
				switch ctyp {
					case GEOMETRY:
						if mesh.geometry != nil {
							return false, errors.New("Invalid mesh")
						}
						mesh.geometry = child.(*Geometry)
					case MATERIAL: 
						mesh.materials.push_back(child.(*Material))
				}
			}
			case SKIN: {
				skin := parent.(*Skin)
				if ctyp == CLUSTER {
					cluster := child.(*Cluster)
					skin.clusters.push_back(cluster)
					if cluster.skin != nil {
						return false, errors.New("Invalid cluster")
					}
					cluster.skin = skin
				}
			}
			case MATERIAL: {
				mat := parent.(*Material)
				if ctyp == TEXTURE {
					ttyp = COUNT
					if con.property == "NormalMap" {
						ttyp = NORMAL
					} else if con.property == "DiffuseColor"
						ttyp = DIFFUSE
					if ttyp == COUNT {
						break
					}
					if mat.textures[ttyp] != nil {
						break
					}
					mat.textures[ttyp] = child.(*Texture)
				}
			}
			case GEOMETRY:
				geom := parent.(*Geometry)
				if ctyp == SKIN { 
					geom.skin = child.(*Skin)
				}
			}
			case CLUSTER:
				cluster := parent.(*Cluster)
				if ctyp == LIMB_NODE || ctyp == MESH || ctyp == NULL_NODE {
					if cluster.link != nil {
						return false, errors.New("Invalid cluster")
					}
					cluster.link = child
				}
			}
			case ANIMATION_LAYER:
				if ctyp == ANIMATION_CURVE_NODE {
					(parent.(*AnimationLayer).curve_nodes.push_back(child.(*AnimationCurveNode))
				}
			}
			case ANIMATION_CURVE_NODE:
				node = parent.(*AnimationCurveNode)
				if ctyp == ANIMATION_CURVE {
					if !node.curves[0].curve == nil {
						node.curves[0].connection = &con
						node.curves[0].curve = child.(*AnimationCurve)
					} else if node.curves[1].curve == nil {
						node.curves[1].connection = &con
						node.curves[1].curve = child.(*AnimationCurve)
					} else if !node.curves[2].curve == nil {
						node.curves[2].connection = &con
						node.curves[2].curve = child.(*AnimationCurve)
					} else {
						return false, errors.New("Invalid animation node")
					}
				}
			}
		}
	}

	for _, iter := range scene.m_object_map {
		obj := iter.second.object
		if obj == nil {
			continue
		}
		if obj.getType() == CLUSTER {
			if !iter.second.object.(*ClusterImpl).postprocess() {
				return false, errors.New("Failed to postprocess cluster")
			}
		}
	}

	return true
}
