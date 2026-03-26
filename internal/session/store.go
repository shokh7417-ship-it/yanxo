package session

import "sync"

type Flow string

const (
	FlowNone          Flow = ""
	FlowTaxiCreate    Flow = "taxi_create"
	FlowServiceCreate Flow = "service_create"
	FlowTaxiSearch    Flow = "taxi_search"
	FlowServiceSearch Flow = "service_search"
)

type Step string

const (
	StepNone Step = ""

	// taxi create
	StepTaxiFromCity      Step = "taxi_from"
	StepTaxiToCity        Step = "taxi_to"
	StepTaxiRideDate      Step = "taxi_date"
	StepTaxiDepartureTime Step = "taxi_time"
	StepTaxiCarType       Step = "taxi_car"
	StepTaxiTotalSeats    Step = "taxi_total"
	StepTaxiContact       Step = "taxi_contact"
	StepTaxiPreview       Step = "taxi_preview"

	// service create
	StepServiceCategory   Step = "service_category"
	StepServicePick       Step = "service_pick"
	StepServiceCustomType Step = "service_custom_type"
	StepServiceArea       Step = "service_area"
	StepServiceNote    Step = "service_note"
	StepServiceContact Step = "service_contact"
	StepServicePreview Step = "service_preview"

	// search
	StepTaxiSearchFrom   Step = "taxi_search_from"
	StepTaxiSearchTo     Step = "taxi_search_to"
	StepServiceSearchCategory   Step = "service_search_category"
	StepServiceSearchPick       Step = "service_search_pick"
	StepServiceSearchCustomType Step = "service_search_custom_type"
	StepServiceSearchArea       Step = "service_search_area"
)

type TaxiDraft struct {
	FromCity      string
	ToCity        string
	RideDate      string
	DepartureTime string
	CarType       string
	TotalSeats    int
	OccupiedSeats int
	Contact       *string
}

type ServiceDraft struct {
	ServiceType string
	// PickCategory: ServicePickCatBuild / auto / wood — faqat StepServicePick paytida.
	PickCategory string
	Area         string
	Note         *string
	Contact      *string
}

type State struct {
	Flow Flow
	Step Step

	Taxi    TaxiDraft
	Service ServiceDraft

	Search SearchDraft
}

type SearchDraft struct {
	TaxiFrom string
	TaxiTo   string

	ServiceType         string
	ServiceArea         string
	ServicePickCategory string // ServicePickCatBuild / auto / wood — StepServiceSearchPick
}

type Store struct {
	mu   sync.RWMutex
	data map[int64]State
}

func NewStore() *Store {
	return &Store{data: map[int64]State{}}
}

func (s *Store) Get(userID int64) (State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[userID]
	return v, ok
}

func (s *Store) Set(userID int64, st State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[userID] = st
}

func (s *Store) Clear(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, userID)
}

