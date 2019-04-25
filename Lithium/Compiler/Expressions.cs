using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using System.Linq.Expressions;
using Expr = System.Linq.Expressions.Expression;

namespace Lithium.Compiler
{
    static class Expressions
    {
        public static Expr EnsureObject(this Expression expression)
        {
            if (expression.Type != typeof(object))
                return Expression.Convert(expression, typeof(object));
            return expression;
        }

        public static Expr MakeGlobalVariableAccessor(ParameterExpression environment, string identifier)
        {
            throw new NotImplementedException();
        }

        public static Expr MakeGlobalVariableWriter(ParameterExpression environment, string identifier, Expression value)
        {
            throw new NotImplementedException();
        }
    }
}
