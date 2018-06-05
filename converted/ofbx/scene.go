package ofbx

import "fmt"

type Connection struct {
	typ      ConnectionType
	from, to uint64
	property string
}

type ConnectionType int

// Connection Types
const (
	OBJECT_OBJECT   ConnectionType = iota
	OBJECT_PROPERTY ConnectionType = iota
)

type ObjectPair struct {
	element *Element
	object  Obj
}

//const GlobalSettings* getGlobalSettings() const override { return &m_settings; }

type Scene struct {
	m_root_element     *Element
	m_root             *Root
	m_scene_frame_rate float32 // = -1
	m_settings         GlobalSettings
	m_object_map       map[uint64]ObjectPair // Slice or map?
	m_all_objects      []Obj
	m_meshes           []*Mesh
	m_animation_stacks []*AnimationStack
	m_connections      []Connection
	m_data             []byte
	m_take_infos       []TakeInfo
}

func (s *Scene) getRootElement() *Element {
	return s.m_root_element
}
func (s *Scene) getRoot() Obj {
	return s.m_root
}
func (s *Scene) getTakeInfo(name string) *TakeInfo {
	for _, info := range s.m_take_infos {
		if info.name.String() == name {
			return &info
		}
	}
	return nil
}
func (s *Scene) getSceneFrameRate() float32 {
	return s.m_scene_frame_rate
}
func (s *Scene) getMesh(index int) *Mesh {
	//assert(index >= 0);
	//assert(index < m_meshes.size());
	return s.m_meshes[index]
}
func (s *Scene) getAnimationStack(index int) *AnimationStack {
	//assert(index >= 0);
	//assert(index < m_animation_stacks.size());
	return s.m_animation_stacks[index]

}
func (s *Scene) getAllObjects() []Obj {
	return s.m_all_objects
}

func Load(data []byte) (*Scene, error) {
	s := &Scene{}
	s.m_data = make([]byte, len(data))
	copy(s.m_data, data)

	fmt.Println("Starting tokenize")
	root, err := tokenize(s.m_data)
	fmt.Println("Tokenize completed", err)
	if err != nil {
		fmt.Println("Starting TokenizeText")
		root, err = tokenizeText(s.m_data)
		fmt.Println("TokenizeText completed")
		if err != nil {
			return nil, err
		}
	}

	s.m_root_element = root
	//assert(scene.m_root_element);

	// This was commented out already I didn't do it
	//if (parseTemplates(*root.getValue()).isError()) return nil
	fmt.Println("Starting parse connection")
	if ok, err := parseConnection(root, s); !ok {
		return nil, err
	}
	fmt.Println("Starting parse takes")
	if ok, err := parseTakes(s); !ok {
		return nil, err
	}
	fmt.Println("Starting parse objects")
	if ok, err := parseObjects(root, s); !ok {
		return nil, err
	}
	fmt.Println("Parsing global settings")
	parseGlobalSettings(root, s)

	return s, nil
}
