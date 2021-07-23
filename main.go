package main

import (
	"errors"
	"fmt"
	"strconv"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
)

// gonum specifics

type State struct {
	Id    int64
	Value interface{}
}

type Link struct {
	Id    int64
	T, F  graph.Node
	Rules map[Operator]Event
}

func (n State) ID() int64 {
	return n.Id
}

func (n State) String() string {
	switch n.Value.(type) {
	case int:
		return strconv.Itoa(n.Value.(int))
	case float32:
		return fmt.Sprintf("%f", n.Value.(float32))
	case float64:
		return fmt.Sprintf("%F", n.Value.(float64))
	case bool:
		return strconv.FormatBool(n.Value.(bool))
	case string:
		return n.Value.(string)
	default:
		return ""
	}
	// return n.Value.(string)
}

func (l Link) From() graph.Node {
	return l.F
}

func (l Link) To() graph.Node {
	return l.T
}

func (l Link) ID() int64 {
	return l.Id
}

func (l Link) ReversedLine() graph.line {
	return Link{F: l.T, T: l.F}
}

// state machine

type Event string
type Operator string

var NodeIDCntr = 0
var LineIDCntr = 1

type StateMachine struct {
	PresentState State
	g            *multi.DirectedGraph
}

func New() *StateMachine {
	s := &StateMachine{}
	s.g = multi.NewDirectionGraph()

	return s
}

func (s *StateMachine) Init(initStateValue interface{}) State {
	s.PresentState = State{Id: int64(NodeIDCntr), Value: initStateValue}
	s.g.AddNode(s.PresentState)
	NodeIDCntr++
	return s.PresentState
}

func (s *StateMachine) NewState(stateValue interface{}) State {
	state := State{Id: int64(NodeIDCntr), Value: stateValue}
	s.g.AddNode(state)
	NodeIDCntr++
	return state
}

func NewRule(triggerConditionOperator Operator, comparisonValue Event) map[Operator]Event {
	return map[Operator]Event{triggerConditionOperator: comparisonValue}
}

func (s *StateMachine) LinkStates(s1, s2 State, rule map[Operator]Event) {
	s.g.SetLine(Link{F: s1, T: s2, Id: int64(LineIDCntr), Rules: rule})
	LineIDCntr++
}

func (s *StateMachine) FireEvent(e Event) error {
	presentNode := s.PresentState

	it := s.g.From(presentNode.Id)

	for it.Next() {
		n := s.g.Node(it.Node().ID()).(State)
		// there can be one defined path between 2 distinct states
		line := graph.LinesOf(s.g.Lines(presentNode.Id, n.Id))[0].(Link)

		for key, val := range line.Rules {
			k := string(key)
			switch k {
			case "eq":
				if val == e {
					s.PresentState = n
					return nil
				}
			default:
				fmt.Printf("Sorry, comparison operator '%s' is not supported\n", k)
				return errors.New("UNSUPPORTED_COMPARISON_OPERATOR")
			}
		}
	}
	return nil
}

func (s *StateMachine) Compute(events []string, printState bool) State {
	for _, e := range events {
		s.FireEvent(Event(e))
		if printState {
			fmt.Printf("%s\n", s.PresentState.String())
		}
	}
	return s.PresentState
}

func main() {
	stateMachine := New()

	initState := stateMachine.Init("locked")
	unlockedState := stateMachine.NewState("unlocked")

	coinRule := NewRule(Operator("eq"), Event("coin"))
	pushRule := NewRule(Operator("eq"), Event("push"))

	stateMachine.LinkStates(initState, unlockedState, coinRule)
	stateMachine.LinkStates(unlockedState, initState, pushRule)

	stateMachine.LinkStates(initState, initState, pushRule)
	stateMachine.LinkStates(unlockedState, unlockedState, coinRule)

	fmt.Printf("Initial state is --------- %s\n", stateMachine.PresentState.String())

	events := []string{"coin", "push"}
	stateMachine.Compute(events, true)

	fmt.Printf("------------- Final state is %s\n", stateMachine.PresentState.String())
}
