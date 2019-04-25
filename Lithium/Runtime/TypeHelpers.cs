using System;
using System.Collections.Generic;
using System.Linq;
using System.Linq.Expressions;
using System.Text;
using System.Threading.Tasks;

namespace Lithium.Runtime
{
    static class TypeHelpers
    {
        public static bool ToBoolean(object value)
        {
            return value is bool ? (bool)value : value != null;
        }

        public static Expression ToBoolean(Expression expr)
        {
            return Expression.Condition(
                Expression.TypeEqual(expr, typeof(bool)),
                Expression.TypeAs(expr, typeof(bool)),
                Expression.Not(Expression.Equal(expr, Expression.Constant(null, typeof(object))))
                );
        }
    }
}
