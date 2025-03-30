package physics

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/google/uuid"
)

// FSS represents a Fluid Simulator Scene file
type FSS struct {
	Units    map[string]*UnitDefinition
	Commands []Command
}

// UnitDefinition represents a unit definition in the FSS file
type UnitDefinition struct {
	Name     string
	Position [3]float64
	Velocity [3]float64
	Mass     float64
	Radius   float64
}

// Command represents a command in the FSS file
type Command struct {
	UnitName string
	Action   string
	Target   string
	Params   []float64
}

// LoadFSSScene loads and executes a FSS file
func (s *Simulation) LoadFSSScene(reader io.Reader) error {
	// Parse FSS file
	fss, err := parseFSS(reader)
	if err != nil {
		return err
	}

	// Clear existing units
	s.ResetSimulation()

	// Create all units first
	units := make(map[string]*Unit)
	for name, def := range fss.Units {
		unit := s.createFSSUnit(def)
		units[name] = unit
		s.Fluid = append(s.Fluid, unit)
		s.Scene.Add(unit.Mesh)
	}

	// Execute all commands
	for _, cmd := range fss.Commands {
		unit, exists := units[cmd.UnitName]
		if !exists {
			return fmt.Errorf("unit not found: %s", cmd.UnitName)
		}

		target, exists := units[cmd.Target]
		if !exists {
			return fmt.Errorf("target unit not found: %s", cmd.Target)
		}

		err := s.executeFSSCommand(unit, target, cmd)
		if err != nil {
			return fmt.Errorf("error executing command: %v", err)
		}
	}

	// Initialize octree with loaded units
	s.updateOctree()

	return nil
}

// parseFSS parses a FSS file and returns a FSS struct
func parseFSS(reader io.Reader) (*FSS, error) {
	fss := &FSS{
		Units:    make(map[string]*UnitDefinition),
		Commands: make([]Command, 0),
	}

	scanner := bufio.NewScanner(reader)
	currentUnit := (*UnitDefinition)(nil)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if line defines a new unit
		if strings.HasSuffix(line, "{") {
			name := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			currentUnit = &UnitDefinition{Name: name}
			continue
		}

		// Check if line ends unit definition
		if line == "}" && currentUnit != nil {
			fss.Units[currentUnit.Name] = currentUnit
			currentUnit = nil
			continue
		}

		// Parse unit properties
		if currentUnit != nil {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}

			property := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch property {
			case "Position":
				fmt.Sscanf(value, "%f, %f, %f", &currentUnit.Position[0], &currentUnit.Position[1], &currentUnit.Position[2])
			case "Velocity":
				fmt.Sscanf(value, "%f, %f, %f", &currentUnit.Velocity[0], &currentUnit.Velocity[1], &currentUnit.Velocity[2])
			case "Mass":
				fmt.Sscanf(value, "%f", &currentUnit.Mass)
			case "Radius":
				fmt.Sscanf(value, "%f", &currentUnit.Radius)
			}
			continue
		}

		// Parse commands
		if strings.Contains(line, ".") && strings.Contains(line, "(") {
			cmd := parseCommand(line)
			if cmd != nil {
				fss.Commands = append(fss.Commands, *cmd)
			}
		}
	}

	return fss, scanner.Err()
}

// parseCommand parses a command line like "Star.orbit(PlanetA, 20)"
func parseCommand(line string) *Command {
	// Remove comments
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)

	// Split into unit and command parts
	parts := strings.Split(line, ".")
	if len(parts) != 2 {
		return nil
	}

	unitName := parts[0]
	cmdPart := parts[1]

	// Extract command name and parameters
	openParen := strings.Index(cmdPart, "(")
	closeParen := strings.Index(cmdPart, ")")
	if openParen == -1 || closeParen == -1 {
		return nil
	}

	action := cmdPart[:openParen]
	paramsStr := cmdPart[openParen+1 : closeParen]
	params := strings.Split(paramsStr, ",")

	if len(params) < 1 {
		return nil
	}

	// First parameter is always the target unit
	target := strings.TrimSpace(params[0])

	// Convert remaining parameters to float64
	var floatParams []float64
	for _, p := range params[1:] {
		var val float64
		fmt.Sscanf(strings.TrimSpace(p), "%f", &val)
		floatParams = append(floatParams, val)
	}

	return &Command{
		UnitName: unitName,
		Action:   action,
		Target:   target,
		Params:   floatParams,
	}
}

// createFSSUnit creates a new unit from a FSS unit definition
func (s *Simulation) createFSSUnit(def *UnitDefinition) *Unit {
	pos := vector3.New(def.Position[0], def.Position[1], def.Position[2])
	vel := vector3.New(def.Velocity[0], def.Velocity[1], def.Velocity[2])
	acc := vector3.Zero[float64]()
	log.Print("unit at position: ", pos)
	// Create unit with Iron as default material
	comp := material.NewComposition(map[*material.Material]float64{
		&material.Iron: 1.0,
	})

	unit := &Unit{
		Id:           uuid.New(),
		_position:    pos,
		_velocity:    vel,
		Acceleration: acc,
		_radius:      def.Radius,
		_mass:        def.Mass,
		Composition:  comp,
		Heat:         0.0,
		canBeAltered: true,
	}

	unit.NewPointLightMesh()
	return unit
}

// executeFSSCommand executes a FSS command
func (s *Simulation) executeFSSCommand(unit *Unit, target *Unit, cmd Command) error {
	switch cmd.Action {
	case "orbit":
		return unit.orbit(target)
	default:
		return fmt.Errorf("unknown command: %s", cmd.Action)
	}
}
