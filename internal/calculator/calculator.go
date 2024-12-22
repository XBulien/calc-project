package calculator

import (
	"errors"
	"strconv"
	"strings"
)

func Calc(expression string) (float64, error) {
	var precedence = map[rune]int{
		'+': 1,
		'-': 1,
		'*': 2,
		'/': 2,
	}

	tokens := tokenize(expression)
	if len(tokens) == 0 {
		return 0, errors.New("empty expression")
	}

	postfix, err := infixToPostfix(tokens, precedence)
	if err != nil {
		return 0, err
	}

	return evaluatePostfix(postfix)
}

func tokenize(expression string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, ch := range expression {
		if ch == ' ' {
			continue
		}
		if isOperator(ch) || ch == '(' || ch == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(ch))
		} else if isDigit(ch) || ch == '.' {
			currentToken.WriteRune(ch)
		} else {
			return nil
		}
	}
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	return tokens
}

func isOperator(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/'
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || ch == '.'
}

func infixToPostfix(tokens []string, precedence map[rune]int) ([]string, error) {
	var output []string
	var stack []rune

	for _, token := range tokens {
		if isDigit(rune(token[0])) {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, '(')
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				output = append(output, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			stack = stack[:len(stack)-1]
		} else if isOperator(rune(token[0])) {
			for len(stack) > 0 && precedence[rune(token[0])] <= precedence[stack[len(stack)-1]] {
				output = append(output, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, rune(token[0]))
		} else {
			return nil, errors.New("invalid token: " + token)
		}
	}

	for len(stack) > 0 {
		output = append(output, string(stack[len(stack)-1]))
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func evaluatePostfix(postfix []string) (float64, error) {
	var stack []float64

	for _, token := range postfix {
		if isDigit(rune(token[0])) {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, num)
		} else if isOperator(rune(token[0])) {
			if len(stack) < 2 {
				return 0, errors.New("not enough operands")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token[0] {
			case '+':
				stack = append(stack, a+b)
			case '-':
				stack = append(stack, a-b)
			case '*':
				stack = append(stack, a*b)
			case '/':
				if b == 0 {
					return 0, errors.New("division by zero")
				}
				stack = append(stack, a/b)
			default:
				return 0, errors.New("invalid operator: " + token)
			}
		} else {
			return 0, errors.New("invalid token: " + token)
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("error in calculation")
	}

	return stack[0], nil
}
