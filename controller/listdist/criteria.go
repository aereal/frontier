package listdist

import (
	"regexp"
	"sync"

	"github.com/aereal/frontier"
)

type CriterionKey string

const (
	CriterionKeyDistributionDomainName CriterionKey = ".Distribution.DomainName"
	CriterionKeyDistributionIsEnabled  CriterionKey = ".Distribution.IsEnabled"
	CriterionKeyEventType              CriterionKey = ".EventType"
	CriterionKeyFunctionArn            CriterionKey = ".Function.ARN"
)

type EqualEventTypeCriterion struct{ EventType string }

var _ Criterion = (*EqualEventTypeCriterion)(nil)

func (EqualEventTypeCriterion) Key() CriterionKey { return CriterionKeyEventType }

func (criterion *EqualEventTypeCriterion) Satisfy(a frontier.FunctionAssociation) bool {
	return a.EventType == criterion.EventType
}

func EqualEventType(eventType string) *EqualEventTypeCriterion {
	return &EqualEventTypeCriterion{EventType: eventType}
}

type EqualDistributionIsEnabledCriterion struct{ IsEnabled bool }

var _ Criterion = (*EqualDistributionIsEnabledCriterion)(nil)

func (EqualDistributionIsEnabledCriterion) Key() CriterionKey {
	return CriterionKeyDistributionIsEnabled
}

func (criterion *EqualDistributionIsEnabledCriterion) Satisfy(a frontier.FunctionAssociation) bool {
	return a.Distribution.IsEnabled == criterion.IsEnabled
}

func EqualDistributionIsEnabled(enabled bool) *EqualDistributionIsEnabledCriterion {
	return &EqualDistributionIsEnabledCriterion{IsEnabled: enabled}
}

type EqualFunctionArnCriterion struct{ FunctionArn string }

var _ Criterion = (*EqualFunctionArnCriterion)(nil)

func (EqualFunctionArnCriterion) Key() CriterionKey { return CriterionKeyFunctionArn }

func (criterion *EqualFunctionArnCriterion) Satisfy(a frontier.FunctionAssociation) bool {
	return a.Function.ARN == criterion.FunctionArn
}

func EqualFunctionArn(functionArn string) *EqualFunctionArnCriterion {
	return &EqualFunctionArnCriterion{FunctionArn: functionArn}
}

type EqualDistributionDomainNameCriterion struct{ DomainName string }

var _ Criterion = (*EqualDistributionDomainNameCriterion)(nil)

func (EqualDistributionDomainNameCriterion) Key() CriterionKey {
	return CriterionKeyDistributionDomainName
}

func (criterion *EqualDistributionDomainNameCriterion) Satisfy(a frontier.FunctionAssociation) bool {
	return a.Distribution.DomainName == criterion.DomainName
}

func EqualDistributionDomainName(domainName string) *EqualDistributionDomainNameCriterion {
	return &EqualDistributionDomainNameCriterion{DomainName: domainName}
}

type MatchDistributionDomainNameCriterion struct{ Pattern *regexp.Regexp }

var _ Criterion = (*MatchDistributionDomainNameCriterion)(nil)

func (MatchDistributionDomainNameCriterion) Key() CriterionKey {
	return CriterionKeyDistributionDomainName
}

func (criterion *MatchDistributionDomainNameCriterion) Satisfy(a frontier.FunctionAssociation) bool {
	return criterion.Pattern.MatchString(a.Distribution.DomainName)
}

func MatchDistributionDomainName(pattern *regexp.Regexp) *MatchDistributionDomainNameCriterion {
	return &MatchDistributionDomainNameCriterion{Pattern: pattern}
}

type Criterion interface {
	Key() CriterionKey
	Satisfy(association frontier.FunctionAssociation) bool
}

func NewCriteria(criterion ...Criterion) *Criteria {
	ret := &Criteria{dirty: map[CriterionKey]Criterion{}}
	for _, c := range criterion {
		ret.unsafeAdd(c)
	}
	return ret
}

type Criteria struct {
	mux   sync.Mutex
	dirty map[CriterionKey]Criterion
}

func (criteria *Criteria) Satisfy(a frontier.FunctionAssociation) bool {
	for _, c := range criteria.dirty {
		if !c.Satisfy(a) {
			return false
		}
	}
	return true
}

func (criteria *Criteria) Add(criterion Criterion) {
	criteria.mux.Lock()
	defer criteria.mux.Unlock()
	criteria.unsafeAdd(criterion)
}

func (criteria *Criteria) unsafeAdd(criterion Criterion) {
	if criteria.dirty == nil {
		criteria.dirty = map[CriterionKey]Criterion{}
	}
	criteria.dirty[criterion.Key()] = criterion
}
