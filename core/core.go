package core

import (
	"fmt"
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
	position := Vector{X: x, Y: y}
	car := &Car{connection: conn}
	car.SetPosition(position)
	car.SetMass(60.0)
	car.SetApplyingForce(Vector{X: 0, Y: -2})
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
				s.computeStep(1)
			}
		}()
	}
}

func (s *Session) computeStep(duration float64) {
	for _, car := range s.Cars {
		car.Integrate(duration)
		car.DoExchange()
	}
}

//Stop function
func (s *Session) Stop() {
	s.status = "stopped"
}

//Car struct
type Car struct {
	position     Vector
	angle        float64
	velocity     Vector
	acceleration Vector
	p1           Vector
	p2           Vector
	p3           Vector
	p4           Vector
	mass         float64
	connection   *websocket.Conn
}

func (c *Car) receiveData() {

}

func (c *Car) sendData() {
	c.connection.WriteJSON(&SentData{
		X:     c.GetPosition().X,
		Y:     c.GetPosition().Y,
		Angle: 0.0,
	})
}

//SentData struct represents sent data
type SentData struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
}

func (c *Car) DoExchange() {
	c.receiveData()
	c.sendData()
}

//SetPosition function
func (c *Car) SetPosition(position Vector) {
	c.position = position
	c.p1 = Vector{X: 2, Y: -3}
	c.p2 = Vector{X: 2, Y: 3}
	c.p3 = Vector{X: 2, Y: 3}
	c.p4 = Vector{X: -2, Y: 3}
	//c.velocity = Vector{X: 0.0, Y: 0.0}
	c.angle = 0.0
}

func (c *Car) GetPosition() Vector {
	return c.position
}

//SetMass function
func (c *Car) SetMass(mass float64) {
	c.mass = mass
}

func (c *Car) setVelocity(vector Vector) {
	//c.velocity.X = x
	//c.velocity.Y = y
}

//SetApplyingForce function
func (c *Car) SetApplyingForce(vector Vector) {
	c.acceleration = Vector{
		X: vector.X / c.mass,
		Y: vector.Y / c.mass,
	}
}

//Integrate function
func (c *Car) Integrate(duration float64) {
	c.position = c.position.Add(c.velocity.Scale(duration))
	c.velocity = c.velocity.Add(c.acceleration.Scale(duration))
	fmt.Printf("x = %f y = %f\n", c.position.X, c.position.Y)
}

//Rotate function
func (c *Car) Rotate(radians float64) {
	c.angle += radians
	//rotate velocity and p1, p2, p3, p4 vectors
	c.p1 = c.p1.Rotate(radians)
	c.p2 = c.p2.Rotate(radians)
	c.p3 = c.p3.Rotate(radians)
	c.p4 = c.p4.Rotate(radians)
}

//Vector struct
type Vector struct {
	X float64
	Y float64
}

//Rotate function initializes new Vector instance, that represents rotated variant of origin vector by radians
func (v *Vector) Rotate(radians float64) Vector {
	return Vector{
		X: math.Cos(radians)*v.X + math.Sin(radians)*v.Y,
		Y: -math.Sin(radians)*v.X + math.Cos(radians)*v.Y,
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
