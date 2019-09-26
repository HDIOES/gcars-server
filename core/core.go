package core

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

//Session struct
type Session struct {
	Cars   []*Car
	status string
}

//AddCar function
func (s *Session) AddCar(x float64, y float64, conn *websocket.Conn) {
	car := &Car{}
	car.Init(x, y, conn)
	s.Cars = append(s.Cars, car)
	if len(s.Cars) > 0 && s.status != "running" {
		s.Run()
	}
}

//Run function
func (s *Session) Run() {
	if strings.Compare(s.status, "running") != 0 {
		s.status = "running"
		go func() {
			for strings.Compare(s.status, "running") == 0 {
				time.Sleep(20 * time.Millisecond)
				s.computeStep(0.02)
			}
		}()
	}
}

func (s *Session) computeStep(duration float64) {
	for _, car := range s.Cars {
		select {
		case signal := <-car.Signal:
			log.Println(signal)
			//apply signal from XBox or PlayStation controller
		default:
		}
		car.Integrate(duration)
		sendData := SentData{
			X:           car.GetPosition().X,
			Y:           car.GetPosition().Y,
			Angle:       car.angle,
			ForcePoint:  car.forcePoint,
			EngineForce: car.engineForce,
		}
		if len(car.Output) == 0 {
			car.Output <- sendData
		}
	}
}

//Stop function
func (s *Session) Stop() {
	s.status = "stopped"
}

//Car struct
type Car struct {
	//linear parameters
	position     Vector
	velocity     Vector
	acceleration Vector

	//car parameters
	engineForce      Vector
	engineForceAngle float64
	forcePoint       Vector

	//angular parameters
	angle               float64
	angularVelocity     float64
	angularAcceleration float64

	//constant parameters
	mass            float64
	height          float64
	width           float64
	momentOfInertia float64

	//tech parameters
	connection *websocket.Conn
	Signal     chan SignalData
	Output     chan SentData
}

//Init function
func (c *Car) Init(x float64, y float64, conn *websocket.Conn) {
	c.connection = conn
	c.angle = 0.0
	c.angularVelocity = 0.0
	c.angularAcceleration = 0.0
	c.mass = 500.0
	c.height = 8.0
	c.width = 4.0
	c.position = Vector{
		X: x,
		Y: y,
	}
	c.velocity = Vector{
		X: 20.0,
		Y: 0.0,
	}
	c.acceleration = Vector{
		X: 0.0,
		Y: 0.0,
	}
	c.engineForce = Vector{
		X: 0.0,
		Y: 10000.0,
	}
	c.forcePoint = Vector{
		X: 0.0,
		Y: -c.width / 2,
	}
	c.momentOfInertia = c.mass * (c.height*c.height + c.width*c.width) / 12
	c.Signal = make(chan SignalData, 1)
	c.Output = make(chan SentData, 1)
	go c.DoExchange()
}

//DoExchange function
func (c *Car) DoExchange() {
	var signal *SignalData
	for {
		if len(c.Signal) == 0 {
			c.connection.ReadJSON(&signal)
			c.Signal <- *signal
		}
		select {
		case out := <-c.Output:
			c.connection.WriteJSON(out)
		default:
		}
	}
}

//SentData struct represents sent data
type SentData struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Angle       float64 `json:"angle"`
	EngineForce Vector  `json:"engineForce"`
	ForcePoint  Vector  `json:"forcePoint"`
}

//SignalData struct
type SignalData struct {
	Angle float64 `json:"angle"`
}

//GetPosition function
func (c *Car) GetPosition() Vector {
	return c.position
}

//SetMass function
func (c *Car) SetMass(mass float64) {
	c.mass = mass
}

//ApplyForce function
func (c *Car) ApplyForce(force Vector, forcePoint Vector) {
	//calculate linear acceleration
	c.acceleration = c.acceleration.Add(Vector{
		X: force.X / c.mass,
		Y: force.Y / c.mass,
	})
	forceMoment := forcePoint.VectorProduct(force)
	c.angularAcceleration = forceMoment / c.momentOfInertia
}

//Integrate function
func (c *Car) Integrate(duration float64) {
	c.ApplyForce(c.engineForce, c.forcePoint)
	c.position = c.position.Add(c.velocity.Scale(duration))
	c.velocity = c.velocity.Add(c.acceleration.Scale(duration))
	//then integrate angular movement
	c.angularVelocity = c.angularVelocity + c.angularAcceleration*duration
	deltaAngle := c.angularVelocity * duration
	c.angle = c.angle + deltaAngle
	if c.angle > 2*math.Pi {
		c.angle = c.angle - 2*math.Pi
	}
	if c.angle < -2*math.Pi {
		c.angle = c.angle + 2*math.Pi
	}
	c.RotateCar(deltaAngle)
	c.acceleration = Vector{
		X: 0,
		Y: 0,
	}
	c.angularAcceleration = 0.0
	//fmt.Printf("x = %f y = %f\n", c.position.X, c.position.Y)
}

//RotateCar function
func (c *Car) RotateCar(radians float64) {
	s := c.forcePoint.Subtract(c.engineForce)
	rotatedS := s.Rotate(radians)
	c.forcePoint = c.forcePoint.Rotate(radians)
	c.engineForce = c.forcePoint.Subtract(rotatedS)
}

//Vector struct
type Vector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

//Rotate function initializes new Vector instance, that represents rotated variant of origin vector by radians
func (v *Vector) Rotate(radians float64) Vector {
	if radians > 0 {
		return Vector{
			X: math.Cos(radians)*v.X + math.Sin(radians)*v.Y,
			Y: -math.Sin(radians)*v.X + math.Cos(radians)*v.Y,
		}
	}
	return Vector{
		X: math.Cos(radians)*v.X - math.Sin(radians)*v.Y,
		Y: math.Sin(radians)*v.X + math.Cos(radians)*v.Y,
	}
}

//Add function initializes new Vector instance, that represents sum of origin and added vectors
func (v *Vector) Add(addedVector Vector) Vector {
	return Vector{X: v.X + addedVector.X, Y: v.Y + addedVector.Y}
}

//Subtract function initializes new Vector instance, that represents subtruction of subtracted vector from origin vector
func (v *Vector) Subtract(subtractedVector Vector) Vector {
	return Vector{X: v.X - subtractedVector.X, Y: v.Y - subtractedVector.Y}
}

//ComponentProduct function initializes new Vector, that represents component product calculated by multiplying component of each vector by component of other vector
func (v *Vector) ComponentProduct(vector Vector) Vector {
	return Vector{X: v.X * vector.X, Y: v.Y * vector.Y}
}

//ScalarProduct function returns value, that represent scalar product both vectors A and B (it can be given by formula Ax*Bx+Ay*By+Az*Bz or |A|*|B|*cos(alfa))
func (v *Vector) ScalarProduct(vector Vector) float64 {
	return vector.X*v.X + vector.Y*v.Y
}

//VectorProduct function returns value, thant represent vector product both vectors A and B (formula Ax*By-Bx*Ay)
func (v *Vector) VectorProduct(vector Vector) float64 {
	return v.X*vector.Y - vector.X*v.Y
}

//Normalize function returns new Vector instance that represents normalized vector, given by formula N = [Ax / length, Ay / length]; where length = sqrt(x^2 + y^2)
func (v *Vector) Normalize() Vector {
	vectorLength := math.Sqrt(v.X*v.X + v.Y*v.Y)
	normalizedX := v.X / vectorLength
	normalizedY := v.Y / vectorLength
	return Vector{X: normalizedX, Y: normalizedY}
}

//Scale function returns new Vector that represents scaled vector based on origin vector, given by formula S = [x * number, y * number]
func (v *Vector) Scale(number float64) Vector {
	return Vector{X: v.X * number, Y: v.Y * number}
}
