package domain

// Token is a task token (rules §7).
type Token string

const (
	TokenNone   Token = ""
	Token1      Token = "1"
	Token2      Token = "2"
	Token3      Token = "3"
	Token4      Token = "4"
	Token5      Token = "5"
	TokenOmega  Token = "omega"  // the last task fulfilled
	TokenArrow1 Token = "arrow1" // ›
	TokenArrow2 Token = "arrow2" // ››
	TokenArrow3 Token = "arrow3" // ›››
	TokenArrow4 Token = "arrow4" // ››››
)

var tokenNumbers = map[Token]int{Token1: 1, Token2: 2, Token3: 3, Token4: 4, Token5: 5}
var tokenArrows = map[Token]int{TokenArrow1: 1, TokenArrow2: 2, TokenArrow3: 3, TokenArrow4: 4}

// Number is k for a numbered token, else 0.
func (t Token) Number() int { return tokenNumbers[t] }

// Arrow is the number of arrows of an arrow token, else 0.
func (t Token) Arrow() int { return tokenArrows[t] }

// Valid reports whether t is a known token or TokenNone.
func (t Token) Valid() bool {
	return t == TokenNone || t == TokenOmega || t.Number() > 0 || t.Arrow() > 0
}

// orderViolation checks the order conditions of the tokens (rules §7, with the generalised
// same-trick rule). tokens[i] is task i's token and at[i] the 1-based trick in which task i was
// fulfilled, 0 while it is open. It returns the index of a task whose condition is broken, or -1.
//
// For a task fulfilled in trick t, b is the number of tasks fulfilled before t and s the number
// fulfilled in t:
//   - numbered k: valid iff b < k ≤ b+s; while open, it fails once k other tasks are fulfilled;
//   - Ω: valid iff no task is open after trick t;
//   - arrows: › before ›› before ››› before ››››, ties (the same trick) allowed; an arrow task
//     fails once a later arrow is fulfilled before it.
func orderViolation(tokens []Token, at []int) int {
	for i, tok := range tokens {
		if k := tok.Number(); k > 0 {
			before, same, done := 0, 0, 0
			for j, a := range at {
				if a == 0 || j == i {
					continue
				}
				done++
				switch {
				case at[i] > 0 && a < at[i]:
					before++
				case a == at[i]:
					same++
				}
			}
			if at[i] == 0 {
				if done >= k {
					return i
				}
			} else if !(before < k && k <= before+same+1) {
				return i
			}
		}
		if tok == TokenOmega && at[i] > 0 {
			for _, a := range at {
				if a == 0 || a > at[i] {
					return i
				}
			}
		}
		if n := tok.Arrow(); n > 0 {
			for j, other := range tokens {
				if other.Arrow() > n && at[j] > 0 && (at[i] == 0 || at[i] > at[j]) {
					return i
				}
			}
		}
	}
	return -1
}
