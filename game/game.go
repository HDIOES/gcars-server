package game

import (
	"log"
	"strings"
	"time"

	"github.com/ByteArena/box2d"
	"github.com/gorilla/websocket"
)

//ServerInstance struct
type ServerInstance struct {
	idCounter int64
	Sessions  []Session
}

//Player struct represents player data and player logic
type Player struct {
	ID           int64
	connection   *websocket.Conn
	bodyDef      *box2d.B2BodyDef
	body         *box2d.B2Body
	shape        *box2d.B2PolygonShape
	sentData     *SentData
	receivedData *ReceivedData
}

//SentData struct represents sent data
type SentData struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
}

//ReceivedData struct represents received data
type ReceivedData struct {
}

func (player *Player) sendData() {
	player.connection.WriteJSON(&SentData{
		X:     player.body.GetPosition().X,
		Y:     player.body.GetPosition().Y,
		Angle: player.body.GetAngle(),
	})
}

func (player *Player) receiveData() {

}

func (player *Player) DoExchange() {
	player.receiveData()
	player.sendData()
}

//Session struct represent game session beetween players
type Session struct {
	ID        int64
	idCounter int64
	Players   []Player
	world     *box2d.B2World
	status    string
}

func (s *Session) CreatePlayer(conn *websocket.Conn) {
	if s.world == nil {
		world := box2d.MakeB2World(box2d.MakeB2Vec2(10.0, 0.0))
		s.world = &world
	}
	s.idCounter++
	player := Player{ID: s.idCounter, connection: conn}
	bodyDef := box2d.MakeB2BodyDef()
	bodyDef.Type = box2d.B2BodyType.B2_dynamicBody
	player.bodyDef = &bodyDef
	player.body = s.world.CreateBody(player.bodyDef)
	player.body.SetTransform(box2d.MakeB2Vec2(500.0, 500.0), 0)
	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(2, 4)
	player.shape = &shape
	player.body.CreateFixture(player.shape, 0.001)
	player.body.SetAngularDamping(0.1)
	s.Players = append(s.Players, player)
	if len(s.Players) >= 1 && s.status != "active" {
		s.StartSession()
	}
}

func (s *Session) StartSession() {
	s.status = "active"
	go s.simulate()
}

func (s *Session) StopSession() {
	s.status = "stopped"
}

func (s *Session) simulate() {
	log.Println("Session started")
	for strings.Compare(s.status, "active") == 0 {
		time.Sleep(20 * time.Millisecond)
		for _, player := range s.Players {
			player.DoExchange()
		}
		s.world.Step(20, 20, 15)
	}
	log.Println("Session stopped")
}

func (si *ServerInstance) CreateSession() {
	si.idCounter++
	session := Session{}
	si.Sessions = append(si.Sessions, session)
}

func CreateServerInstance() *ServerInstance {
	serverInstance := ServerInstance{}
	return &serverInstance
}
