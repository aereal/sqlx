package parser

import (
	"strings"

	"github.com/aereal/sqlx/models"
	"github.com/aereal/sqlx/sql/ast"
)

func (p *Parser) parseSetStatement() (*ast.SetStatement, error) {
	stmt := &ast.SetStatement{
		Lifetime: ast.SetStatementLifetimeSession,
	}
	if p.isTokenMatch("LOCAL") {
		p.advance()
		stmt.Lifetime = ast.SetStatementLifetimeLocal
	}

	name := p.parseIdent()
	if name == nil {
		return nil, p.expectedError("identifier")
	}
	stmt.ParameterName = name.Name

	if !p.matchType(models.TokenTypeEq) {
		return nil, p.expectedError("=")
	}

	switch {
	case p.isStringLiteral():
		stmt.ParameterValue = &ast.LiteralValue{Type: "string", Value: p.currentToken.Token.Value}
		p.advance()
	case p.isBooleanLiteral():
		stmt.ParameterValue = &ast.LiteralValue{Type: "bool", Value: p.currentToken.Token.Value}
		p.advance()
	case p.isNumericLiteral():
		lit := &ast.LiteralValue{
			Type:  "int",
			Value: p.currentToken.Token.Value,
		}
		if strings.ContainsAny(p.currentToken.Token.Value, ".eE") {
			lit.Type = "float"
		}
		stmt.ParameterValue = lit
		p.advance()
	case p.currentToken.Token.Type == models.TokenTypeOn:
		stmt.ParameterValue = &ast.LiteralValue{Type: "bool", Value: "on"}
		p.advance()
	case p.isIdentifier():
		id := p.parseIdentAsString()
		lit := &ast.LiteralValue{Type: "string", Value: id}
		switch strings.ToLower(id) {
		case "yes", "no", "off": // `on` is handled above case
			lit.Type = "bool"
		}
		stmt.ParameterValue = lit
	default:
		return nil, p.expectedError("expression")
	}

	return stmt, nil
}
