package parser

// https://www.meziantou.net/creating-a-parser-for-boolean-expressions.htm
// https://stackoverflow.com/questions/17568067/how-to-parse-a-boolean-expression-and-load-it-into-a-class
// https://gist.github.com/leehsueh/1290686/36b0baa053072c377ac7fc801d53200d17039674
// https://unnikked.ga/how-to-build-a-boolean-expression-evaluator-518e9e068a65

type Expr struct {
	Op    string
	Left  *Expr
	Right *Expr
	Val   string
}

type Parser struct {
	lexer Lexer
}

func NewParser(lexer Lexer) Parser {
	return Parser{
		lexer: lexer,
	}
}

func (p *Parser) Parse() *Expr {
	tokens := p.lexer.Tokens()

	for _, t := range tokens {
		p := Expr{
			Op:    t.Type,
			Left:  nil,
			Right: nil,
			Val:   t.Value,
		}

		return &p
	}

	return nil
}
