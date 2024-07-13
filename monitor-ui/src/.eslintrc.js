module.exports = {
    "ignorePatterns": [
        "node_modules/"
    ],
    "rules": {
        "comma-dangle": 0,
        "eqeqeq": [
            2,
            "always"
        ],
        "indent": [
            2,
            2,
            {
                "VariableDeclarator": 1,
                "SwitchCase": 1
            }
        ],
        "no-useless-escape": 0,
        "quotes": [
            2,
            "single"
        ],
        "semi": ['error', 'never'],
        "no-extra-semi": 0,
        "spaced-comment": [
            2,
            "always",
            {
                "markers": [
                    "/"
                ]
            }
        ],
        "space-infix-ops": 0,
        "no-async-promise-executor": 0,
        "space-before-function-paren": [
            2,
            {
                "anonymous": "always",
                "named": "never",
                "asyncArrow": "always"
            }
        ],
        "key-spacing": [
            2,
            {
                "beforeColon": false,
                "afterColon": true
            }
        ],
        "keyword-spacing": [
            2,
            {
                "before": true,
                "after": true
            }
        ],
        "comma-spacing": 0,
        "space-in-parens": [
            2,
            "never"
        ],
        "arrow-spacing": [
            2,
            {
                "before": true,
                "after": true
            }
        ],
        "arrow-parens": [
            2,
            "as-needed"
        ],
        "max-depth": [
            2,
            8
        ],
        "handle-callback-err": [
            2,
            "^(err|error)$"
        ],
        "no-eval": 2,
        "no-with": 2,
        "no-extra-boolean-cast": 2,
        "no-trailing-spaces": 2,
        "no-spaced-func": 2,
        "semi-spacing": [
            2,
            {
                "before": false,
                "after": true
            }
        ],
        "space-before-blocks": 0,
        "no-param-reassign": [
            2,
            {
                "props": false
            }
        ],
        "no-multiple-empty-lines": [
            2,
            {
                "max": 1,
                "maxBOF": 0,
                "maxEOF": 0
            }
        ],
        "no-var": 2,
        "no-multi-spaces": 2,
        "operator-linebreak": [
            "error",
            "before",
            {
                "overrides": {
                    "=": "none"
                }
            }
        ],
        "max-params": [
            2,
            5
        ],
        "max-lines": [
            2,
            {
                "max": 4000,
                "skipBlankLines": true,
                "skipComments": true
            }
        ],
        "no-constant-condition": 2,
        "object-curly-spacing": 0,
        "array-bracket-spacing": [
            2,
            "never"
        ],
        "curly": [
            2,
            "all"
        ],
        "eol-last": [
            2,
            "always"
        ],
        "prefer-const": 2,
        "no-else-return": 2,
        "brace-style": [
            1,
            "stroustrup",
            {}
        ],
        "newline-per-chained-call": [
            "error",
            {
                "ignoreChainWithDepth": 2
            }
        ],
        "object-shorthand": [
            2,
            "always",
            {
                "ignoreConstructors": false,
                "avoidQuotes": true
            }
        ],
        "array-callback-return": 2,
        "object-curly-newline": [
            2,
            {
                "ObjectExpression": {
                    "minProperties": 4,
                    "multiline": true,
                    "consistent": true
                },
                "ObjectPattern": {
                    "minProperties": 4,
                    "multiline": true,
                    "consistent": true
                },
                "ImportDeclaration": {
                    "minProperties": 4,
                    "multiline": true,
                    "consistent": true
                },
                "ExportDeclaration": {
                    "minProperties": 4,
                    "multiline": true,
                    "consistent": true
                }
            }
        ],
        "function-call-argument-newline": [
            2,
            "consistent"
        ],
        "function-paren-newline": [
            2,
            "multiline-arguments"
        ],
        "object-property-newline": [
            2,
            {
                "allowAllPropertiesOnSameLine": false
            }
        ],
        "quote-props": [
            2,
            "as-needed",
            {
                "keywords": false,
                "unnecessary": true,
                "numbers": false
            }
        ],
        "array-bracket-newline": [
            "off",
            "consistent"
        ],
        "array-element-newline": [
            "off",
            {
                "multiline": true,
                "minItems": 3
            }
        ],
        "no-console": [
            "error",
            {
                "allow": [
                    "error",
                    "warn",
                    "info"
                ]
            }
        ],
        "arrow-body-style": [
            "error",
            "as-needed",
            {
                "requireReturnForObjectLiteral": false
            }
        ],
        "vue/html-indent": [
            2,
            2
        ],
        "vue/html-closing-bracket-newline": [
            2,
            {
                "singleline": "never",
                "multiline": "always"
            }
        ],
        "vue/html-self-closing": [
            2,
            {
                "html": {
                    "void": "always"
                }
            }
        ],
        "vue/multi-word-component-names": 0,
        "vue/component-name-in-template-casing": [
            2,
            "kebab-case",
            {
                "registeredComponentsOnly": false
            }
        ],
        "vue/object-curly-spacing": 2,
        "vue/space-infix-ops": 2,
        "vue/key-spacing": [
            2,
            {
                "beforeColon": false,
                "afterColon": true
            }
        ],
        "vue/eqeqeq": 2,
        "vue/comma-dangle": 2,
        "vue/array-bracket-spacing": [
            2,
            "never"
        ]
    }
}