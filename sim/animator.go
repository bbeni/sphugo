package sim

import (
	"image"
	"math"
	"log"
	"fmt"
	"os"
	"strings"
	"image/png"
	"image/draw"

	"github.com/bbeni/sphugo/gx"
	"github.com/go-gl/gl/v4.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Animator struct {

	//frames for rendering
	Frames   []image.Image
	//FramesMu sync.Mutex

	// as refernce to simulation variables
	Simulation *Simulation

	// reference to particles
	// also used to order particles acoording to z-value before rendering
	renderingParticleArray []*Particle
}

func MakeAnimator(simulation *Simulation) Animator {

	if simulation.Root == nil || len(simulation.Root.Particles) == 0  {
		panic("int Run(): Simulation not initialized!")
	}

	ani := Animator {
		renderingParticleArray: make([]*Particle, len(simulation.Root.Particles)),
	}

	for i, _ := range simulation.Root.Particles {
		ani.renderingParticleArray[i] = &simulation.Root.Particles[i]
	}

	ani.Simulation = simulation
	ani.Frames = make([]image.Image, 0, simulation.Config.NSteps)

	// render first frame
	ani.Frame()
	return ani
}


func (ani *Animator) CurrentFrame() gx.Canvas {
	//
	// order according to z-value
	//

	ani.renderingParticleArray = make([]*Particle, len(ani.Simulation.Root.Particles))

	for i, _ := range ani.Simulation.Root.Particles {
		ani.renderingParticleArray[i] = &ani.Simulation.Root.Particles[i]
	}

	extractZindex := func(p *Particle) int {
		//return int(p.Rho*100000)
		return -p.Z
	}

	QuickSort(ani.renderingParticleArray, extractZindex)

	canvas := gx.NewCanvas(1280, 720)
	canvas.Clear(gx.BLACK)

	for _, particle := range ani.renderingParticleArray {
		x := float32(particle.Pos.X) * float32(canvas.W)
		y := float32(particle.Pos.Y) * float32(canvas.H)

		//zNormalized := float32(particle.Z)/float32(math.MaxInt)
		//color_index := 255 - uint8((particle.Rho - 1)*64)
		//color_index := 255 - uint8(zNormalized * 256)

		m := ani.Simulation.Config.ParticleMass
		colorFormula := float64(particle.Rho/(m* float64(len(ani.Simulation.Root.Particles)*10))*256)
		//colorFormula := float64(particle.Vel.Norm()*256)

		color_index := uint8(math.Min(colorFormula, 255))
		color := gx.ParaRamp(color_index)
		//color := gx.HeatRamp(color_index)
		//color := gx.ToxicRamp(color_index)
	    //color := gx.RainbowRamp(255 - color_index)


		if color_index > 255 {
			nnRadius := float32(particle.NNDists[0])*float32(canvas.W)
			canvas.DrawCircle(x, y, nnRadius, 2, gx.WHITE)
		}

		//canvas.DrawDisk(float32(x), float32(y), zNormalized*zNormalized*20+1, color)
		canvas.DrawDisk(float32(x), float32(y), 4, color)
	}
	return canvas
}

// save the last frame of the simultion in the buffer
func (ani *Animator) Frame() {
	canvas := ani.CurrentFrame()
	ani.Frames = append(ani.Frames, canvas.Img)
}

func (ani *Animator) FrameToPNG(file_path string, i int) bool {

	// TODO: check i for bounds

	file, err := os.Create(file_path)
	defer file.Close()

	if err != nil {
		fmt.Printf("Error: couldn't create file %q : %q", file_path, err)
		return false
	}

	err = png.Encode(file, ani.Frames[i])
	if err != nil {
		log.Printf("Error: couldn't encode PNG %q : %q", file_path, err)
		return false
	}

	return true
}


type Frame struct {
	Positions [ ][2]float32
	NNPos	  [2][2]float32
	Densities [ ]float32
}

type AnimatorGL struct {
	sim *Simulation
	Frames []Frame
}

func MakeAnimatorGL(sim *Simulation) AnimatorGL {
	ani := AnimatorGL{
		sim: sim,
		Frames: make([]Frame, 0, sim.Config.NSteps),
	}

	ani.AddFrame()

	return ani
}

func (ani *AnimatorGL) NumberFrames() int {
	return len(ani.Frames)
}

// Adds the current state of the simulation to Frames
func (ani *AnimatorGL) AddFrame() {
	n := len(ani.sim.Root.Particles)

	frame := Frame{
		Positions: make([][2]float32, n),
		Densities: make([]float32, n),
	}

	for i := range n {
		p := &ani.sim.Root.Particles[i]
		frame.Positions[i] = [2]float32{float32(p.Pos.X), float32(p.Pos.Y)}
		frame.Densities[i] = float32(ani.sim.Root.Particles[i].Rho)
		frame.NNPos[0] = [2]float32{float32(p.NNPos[0].X), float32(p.NNPos[0].Y)}
		frame.NNPos[1] = [2]float32{float32(p.NNPos[1].X), float32(p.NNPos[1].Y)}
	}

	ani.Frames = append(ani.Frames, frame)
}

// gl stuff

var previousTime float64
var angle		 float64

var program 	 uint32
var vao 		 uint32
var texture      uint32

var positionUniform int32
var densityUniform int32

var camera        mgl32.Mat4
var cameraUniform int32

func (ani *AnimatorGL) Init(windowWidth, windowHeight int) {

	var err error
	program, err = newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 10.0)
	camera = mgl32.LookAtV(mgl32.Vec3{0.5, 0.5, 1.5}, mgl32.Vec3{0.5, 0.5, 0}, mgl32.Vec3{0, 1, 0})

	projectionUniform := gl.GetUniformLocation(program, gl.Str("Projection\x00"))
	cameraUniform = gl.GetUniformLocation(program, gl.Str("Camera\x00"))
	textureUniform := gl.GetUniformLocation(program, gl.Str("Texture\x00"))
	positionUniform = gl.GetUniformLocation(program, gl.Str("Position\x00"))
	densityUniform = gl.GetUniformLocation(program, gl.Str("Density\x00"))

	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])
	gl.Uniform2f(positionUniform, 0, 0)
	gl.Uniform1i(textureUniform, 0)
	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Load the texture
	texture, err = newTexture("water_droplet.png")
//	texture, err = newTexture("other_drop.png")
	if err != nil {
		log.Fatalln(err)
	}

	// Configure the vertex data
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	//Square
	// x, y, u, v
	quad := []float32{
		-0.5,  0.5,  0,  1,
		0.5,  -0.5,  1,  0,
		-0.5, -0.5,  0,  0,
		-0.5,  0.5,  0,  1,
		0.5,   0.5,  1,  1,
		0.5,  -0.5,  1,  0,
	}

	gl.BufferData(gl.ARRAY_BUFFER, len(quad)*4, gl.Ptr(quad), gl.STATIC_DRAW)

	VertexAttrib := uint32(gl.GetAttribLocation(program, gl.Str("Vertex\x00")))
	gl.EnableVertexAttribArray(VertexAttrib)
	gl.VertexAttribPointerWithOffset(VertexAttrib, 2, gl.FLOAT, false, 4*4, 0)

	UVCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("UVCoord\x00")))
	gl.EnableVertexAttribArray(UVCoordAttrib)
	gl.VertexAttribPointerWithOffset(UVCoordAttrib, 2, gl.FLOAT, false, 4*4, 2*4)

	angle = 0.0
	previousTime = glfw.GetTime()

}

func (ani *AnimatorGL) DrawFrame(index int) {

	fmt.Println(index, len(ani.Frames))

	if index >= len(ani.Frames) {
		index = 0
		//panic("not generated frame yet!")
	}

	gl.UseProgram(program)
	//gl.Enable(gl.BLEND)
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE) // glowy
	//gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.ClearColor(1.0, 0.1, 0.1, 1.0)

	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor(0, 242, 1280, 720)
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
	gl.Disable(gl.SCISSOR_TEST)

	// Update
	time := glfw.GetTime()
	elapsed := time - previousTime
	previousTime = time

	angle += elapsed

	//var x float32= 1.5*float32(math.Sin(angle*13/10))
	//var y float32= 0.9//3*float32(math.Cos(angle))
	//var z float32= 1.9*float32(math.Cos(angle))

	//camera = mgl32.LookAtV(mgl32.Vec3{x, y, z}, mgl32.Vec3{0.5, 0.5, 0}, mgl32.Vec3{0, 1, 0})

	// Render

	gl.BindVertexArray(vao)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	fmt.Println(1.0/elapsed)

	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	frame := &ani.Frames[index]

	for i := range frame.Positions {
		gl.Uniform2f(positionUniform, float32(frame.Positions[i][0]*1.75 - 0.5), float32(1-frame.Positions[i][1]))
		gl.Uniform1f(densityUniform, float32(frame.Densities[i]/10000))
		fmt.Println(float32(frame.Positions[i][0]*1.75 - 0.5))
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

var vertexShader = `
#version 330

uniform mat4 Projection;
uniform mat4 Camera;
uniform vec2 Position;
uniform float Density;


in vec3 Vertex;
in vec2 UVCoord;

out vec2 FragUVCoord;
out vec4 FragColor;
out float FragDensity;


void main() {
    float scale = 0.04f;
    FragUVCoord = UVCoord;
    FragDensity = Density;
   	gl_Position = Projection * Camera * vec4((Vertex.xy  * scale) + Position, 0, 1);
}
` + "\x00"

var fragmentShader = `
#version 330

uniform sampler2D Texture;

in vec2 FragUVCoord;
in float FragDensity;

out vec4 outputColor;

void main() {
	vec4 color = texture(Texture, FragUVCoord);
    //outputColor = vec4((1-FragDensity)*0.34, (1-FragDensity)*0.45, color.z+(1-FragDensity)*0.1, color.w*0.7);
    outputColor = vec4(0.0, 0.0, 1.0, 1.0);
}
` + "\x00"
