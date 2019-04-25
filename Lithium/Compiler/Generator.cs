using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using System.Linq.Expressions;
using Expr = System.Linq.Expressions.Expression;

using Irony.Ast;
using Irony.Parsing;
using Lithium.Exceptions;
using Lithium.Runtime;

namespace Lithium.Compiler
{
    class Generator
    {
        public CompilationScope CurrentScope
        { get; set; }

        public SourceSpan CurrentSpan
        { get; set; }

        public Expr CompileBlock(ParseTreeNode parseNode)
        {
            using (new BlockCompilationScope(this, parseNode.Span))
            {
                return Expr.Block(
                    typeof(object),
                    CurrentScope.Variables,
                    parseNode.ChildNodes.Select(CompileStatement).Where(x => x != null)
                );
            }
        }

        public Expr CompileStatement(ParseTreeNode parseNode)
        {
            switch (parseNode.Term.Name)
            {
                case "GlobalAssignment":
                case "LocalAssignment":
                case "FunctionCall":
                    throw new NotImplementedException();
                case "Label":
                    return CurrentScope.RegisterJumpTarget(parseNode.ChildNodes[0].Token.Text);
                case "Goto":
                    return Expr.Goto(CurrentScope.GetJumpLabel(parseNode.ChildNodes[0].Token.Text));
                case "Break":
                    return Expr.Break(CurrentScope.GetExitLabel());
                case "Scope":
                    return CompileBlock(parseNode);
                case "ForLoop_Enumerable":
                case "ForLoop_Iterable":
                case "WhileLoop":
                case "RepeatLoop":
                case "IfStatement":
                case "GlobalFunction":
                    throw new NotImplementedException();
                case "LocalFunction":
                    return CompileLocalFunction(parseNode);
                case "ReturnStatement":
                    return Expr.Return(CurrentScope.GetReturnLabel(), GetMultiExpressionValue(parseNode.ChildNodes[0]));
                    
                default:
                    throw new LithiumCompilerException("Unrecognized statement terminal '{0}' at {1}", parseNode.Term.Name, parseNode.Span.Location.ToUiString());
            }
        }

        public Expr CompileExpression(ParseTreeNode parseNode)
        {
            switch(parseNode.Term.Name)
            {
                case "Nil":
                    return Expr.Constant(null, typeof(object));
                case "True":
                    return Expr.Constant(true, typeof(bool));
                case "False":
                    return Expr.Constant(false, typeof(bool));
                case "Number":
                    return Expr.Constant(parseNode.Token.Value, typeof(double));
                case "String":
                    return Expr.Constant(parseNode.Token.Value, typeof(string));
                case "Variable":
                    return GetVariable(parseNode);

                default:
                    throw new LithiumCompilerException("Unrecognized expression terminal '{0}' at {1}", parseNode.Term.Name, parseNode.Span.Location.ToUiString());
            }
        }

        #region Functions

        public Expr CompileLocalFunction(ParseTreeNode parseNode)
        {
            var name = parseNode.ChildNodes[0].Token.Text;

            using (var fs = new FunctionScope(this, parseNode.Span, name))
            {
                var parameters = parseNode.ChildNodes[1].ChildNodes.Select(x => GetFunctionParameter(fs, x));

                var function = Expr.Lambda(CompileBlock(parseNode.ChildNodes[2]), name, parameters);
                var targetLocal = fs.ParentScope.RegisterLocal(name);

                return Expr.Assign(targetLocal, function);
            }
        }

        #endregion

        #region Conditionals

        public Expr CompileIfStatement(ParseTreeNode parseNode)
        {
            //[0] - Condition
            //[1] - True Block
            //[2] - {ElseIf Statements}
            //[3] - [Else Block]

            bool hasElseIfs = parseNode.ChildNodes[2].ChildNodes.Any();
            bool hasElse = parseNode.ChildNodes[3].ChildNodes.Any();

            Expr ifStatement = Expr.Empty();

            //Add else component
            if (hasElse) ifStatement = CompileBlock(parseNode.ChildNodes[3]);

            foreach (var elseIf in Enumerable.Reverse(parseNode.ChildNodes[2].ChildNodes))
                ifStatement = Expr.IfThenElse(
                    TypeHelpers.ToBoolean(CompileExpression(elseIf.ChildNodes[0])), 
                    CompileBlock(elseIf.ChildNodes[1]),
                    ifStatement);

            ifStatement = Expr.IfThenElse(
                TypeHelpers.ToBoolean(CompileExpression(parseNode.ChildNodes[0])), 
                    CompileBlock(parseNode.ChildNodes[1]),
                    ifStatement
                );

            return ifStatement;
        }

        #endregion

        #region Private Helpers

        Expr GetVariable(ParseTreeNode variableNode)
        {
            //Doesn't represent a single level variable, but rather a "path" with which
            //to access a variable.
            throw new NotImplementedException();
        }

        /// <summary>
        /// Gets an expression representing the value of one or more expressions, as used by a return statement
        /// </summary>
        Expr GetMultiExpressionValue(ParseTreeNode expressionList)
        {
            if (expressionList.ChildNodes.Count == 0)
                return Expr.Constant(null, typeof(object));
            if (expressionList.ChildNodes.Count == 1)
                return CompileExpression(expressionList.ChildNodes[0]);
            else
                return Expr.NewArrayInit(typeof(object), expressionList.ChildNodes.Select(CompileExpression));
        }

        ParameterExpression GetFunctionParameter(FunctionScope scope, ParseTreeNode parameterNode)
        {
            return scope.RegisterParameter(parameterNode.Token.Text);
        }

        #endregion
    }
}
