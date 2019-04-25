using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using Irony.Parsing;

namespace Lithium.Compiler
{
    [Language("Lua", "5.2", "Lua 5.2 grammar used for Lithium")]
    class Language : Grammar
    {
        public Language()
        {
            var singleLineComment = new CommentTerminal("Comment", "--", "\r", "\n", "\u2085", "\u2028", "\u2029");
            var multilineComment = new LuaLongCommentTerminal("Comment");

            var normalString = new LuaStringLiteral("String");

            var number = new NumberLiteral("Number", NumberOptions.AllowStartEndDot);
            number.AddPrefix("0x", NumberOptions.Hex);
            number.AddExponentSymbols("eE", TypeCode.Double);
            number.DefaultIntTypes = new TypeCode[] { TypeCode.Double };
            number.DefaultFloatType =  TypeCode.Double;

            var name = new IdentifierTerminal("Name");

            var nil = new KeyTerm("nil", "Nil");
            var @true = new KeyTerm("true", "True");
            var @false = new KeyTerm("false", "False");

            var namelist = new NonTerminal("NameList");

            var chunk = new NonTerminal("Chunk");
            var block = new NonTerminal("Block");
            var stat = new NonTerminal("Statement");
            var retstat = new NonTerminal("Return");

            var label = new NonTerminal("Label");
            
            var funcname = new NonTerminal("FunctionName");
            var funcname_normal = new NonTerminal("FunctionName_Normal");
            var funcname_self = new NonTerminal("FunctionName_Self");

            var variables = new NonTerminal("Variables");
            var variable = new NonTerminal("Variable");
            var property = new NonTerminal("Property");
            var index = new NonTerminal("Index");
            
            var explist = new NonTerminal("ExpressionList");
            var expression = new NonTerminal("Expression");

            var binop = new NonTerminal("BinaryOperation");
            var binoperator = new NonTerminal("BinaryOperator");
            var unop = new NonTerminal("UnaryOperation");
            var unoperator = new NonTerminal("UnaryOperator");

            var functiondef = new NonTerminal("FunctionDefinition");
            var tableconstructor = new NonTerminal("TableConstructor");
            var functioncall = new NonTerminal("FunctionCall");
            var args = new NonTerminal("Arguments");
            var funcbody = new NonTerminal("FunctionBody");
            var parlist = new NonTerminal("ParameterList");

            var field = new NonTerminal("Field");
            var fieldsep = new NonTerminal("FieldSeperator");
            var fieldlist = new NonTerminal("FieldList");


            var global_assignment = new NonTerminal("GlobalAssignment");
            var local_assignment = new NonTerminal("LocalAssignment");
            var @break = new NonTerminal("Break");
            var @goto = new NonTerminal("Goto");
            var scope = new NonTerminal("Scope");
            var forLoopEnumerable = new NonTerminal("ForLoop_Enumerable");
            var forLoopIterable = new NonTerminal("ForLoop_Iterable");
            var whileLoop = new NonTerminal("WhileLoop");
            var repeatLoop = new NonTerminal("RepeatLoop");
            var ifStatement = new NonTerminal("IfStatement");
            var elseifStatement = new NonTerminal("ElseIfStatement");
            var elseifStatements = new NonTerminal("ElseIfStatements");
            var elseStatement = new NonTerminal("ElseStatement");
            var globalFunction = new NonTerminal("GlobalFunction");
            var localFunction = new NonTerminal("LocalFunction");
            var @return = new NonTerminal("ReturnStatement");

            var assignmentValues = new NonTerminal("AssignmentValues");
            var localAssignmentValues = new NonTerminal("LocalAssignmentValues");

            var functioncall_normal = new NonTerminal("FunctionCall_Normal");
            var functioncall_instance = new NonTerminal("FunctionCall_Instance");

            MarkTransient(expression, stat, localAssignmentValues, binoperator, unoperator, assignmentValues, chunk);
            
            MarkMemberSelect(".", ":", "[");
            MarkReservedWords("if", "for", "while", "repeat", "do", "end", "break", "return", "local", "function", "false", "true", "nil");
            MarkPunctuation("=", "'", "\"", ".", ":", "::", "[", "]", "(", ")", "return", "if", "for", "then", "else", "elseif", "while", "repeat", "do", "end", "local", "function", "goto");
            RegisterBracePair("[", "]");
            RegisterBracePair("[[", "]]");
            RegisterBracePair("[=[", "]=]");
            RegisterBracePair("[==[", "]==]");
            RegisterBracePair("[===[", "]===]");
            RegisterBracePair("{", "}");
            RegisterBracePair("(", ")");


            chunk.Rule = block;
            block.Rule = MakeStarRule(block, stat);

            stat.Rule = ToTerm(";")
                | global_assignment 
                | local_assignment
                | functioncall
                | label
                | @break
                | @goto
                | scope
                | forLoopEnumerable
                | forLoopIterable
                | whileLoop
                | repeatLoop
                | ifStatement
                | globalFunction
                | localFunction
                | retstat
                | singleLineComment
                | multilineComment
                ;

            assignmentValues.Rule = ToTerm("=") + explist;
            localAssignmentValues.Rule = ToTerm("=") + explist | Empty;


            global_assignment.Rule = variables + assignmentValues;
            local_assignment.Rule = ToTerm("local") + variables + localAssignmentValues;
            @break.Rule = ToTerm("break");
            @goto.Rule = ToTerm("goto") + name;
            scope.Rule = ToTerm("do") + block + "end";
            forLoopIterable.Rule = ToTerm("for") + name + "=" + expression + "," + expression + ("," + expression).Q() + "do" + block + "end";
            forLoopEnumerable.Rule = ToTerm("for") + variables + "in" + explist + "do" + block + "end";
            whileLoop.Rule = ToTerm("while") + expression + "do" + block + "end";
            repeatLoop.Rule = ToTerm("repeat") + block + "until" + expression;
            ifStatement.Rule = ToTerm("if") + expression + "then" + block + elseifStatements + elseStatement + "end";
            elseifStatement.Rule = ToTerm("elseif") + expression + "then" + block;
            elseifStatements.Rule = MakeStarRule(elseifStatements, elseifStatement);
            elseStatement.Rule = "else" + block | Empty;
            globalFunction.Rule = ToTerm("function") + funcname + funcbody;
            localFunction.Rule = ToTerm("local") + "function" + name + funcbody;

            retstat.Rule = ToTerm("return") + explist;
            label.Rule = ToTerm("::") + name + "::";
            namelist.Rule = MakePlusRule(namelist, ToTerm(","), name);

            funcname.Rule = funcname_normal | funcname_self;
            funcname_normal.Rule = MakePlusRule(funcname_normal, ToTerm("."), name);
            funcname_self.Rule = funcname + ":" + name;

            explist.Rule = MakeListRule(explist, ToTerm(","), expression, TermListOptions.StarList);

            expression.Rule = PreferShiftHere() + nil
                | @true
                | @false
                | number
                | variable
                | functioncall
                | normalString
                | "..."
                | functiondef
                | tableconstructor
                | binop
                | unop;
            expression.ErrorAlias = "valid expression";

            functioncall.Rule = functioncall_normal | functioncall_instance;


            //LALR supporting prefixes
            variables.Rule = MakeListRule(variables, ToTerm(","), variable, TermListOptions.PlusList);
            property.Rule = variable + "." + name | functioncall + "." + name;
            index.Rule = variable + "[" + expression + "]" | functioncall + "[" + expression + "]";
            variable.Rule = name | property | index | "(" + expression + ")";


            functioncall_instance.Rule = variable + ":" + name + args | functioncall + ":" + name + args;
            functioncall_normal.Rule = variable + args | functioncall + args;

            args.Rule = ToTerm("(") + explist + ")" | tableconstructor | normalString;
            
            functiondef.Rule = ToTerm("function") + funcbody;
            funcbody.Rule = ToTerm("(") + parlist + ")" + block + "end";

            parlist.Rule = namelist | namelist + "," + "..." | "..." | Empty;
            tableconstructor.Rule = ToTerm("{") + fieldlist + "}";

            fieldsep.Rule = ToTerm(",") | ";";
            //WARNING: Doesn't allow trailing delimiter ( { field1, field2, } )
            fieldlist.Rule = MakeListRule(fieldlist, fieldsep, field, TermListOptions.StarList | TermListOptions.AddPreferShiftHint);
            field.Rule = (ToTerm("[") + expression + "]" + "=" + expression) | (name + "=" + expression) | expression;


            var binMinus = ToTerm("-");
            var unMinus = ToTerm("-");

            binoperator.Rule = ToTerm("+") | binMinus | "*" | "/" | "^" | "%" | ".." | "<" | "<=" | ">" | ">=" | "==" | "~=" | "and" | "or";
            unoperator.Rule = unMinus | "#" | "not";
            unoperator.SetFlag(TermFlags.IsOperator);
            binoperator.SetFlag(TermFlags.IsOperator);


            RegisterOperators(5, Associativity.Left, "^", "%", "..");
            RegisterOperators(4, Associativity.Left, "/");
            RegisterOperators(2, Associativity.Left, "*");
            RegisterOperators(2, Associativity.Right, unMinus);
            RegisterOperators(1, Associativity.Left, "+");
            RegisterOperators(0, Associativity.Left, binMinus);
            RegisterOperators(-1, Associativity.Left, "<", "<=", ">", ">=", "==", "~=", "and", "or");

            binop.Rule = expression + binoperator + expression;
            unop.Rule = ReduceHere() + unoperator + expression;

            this.Root = chunk;
        }
    }
}
