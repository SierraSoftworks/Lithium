using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Lithium.Exceptions
{
    public class LithiumCompilerException : LithiumException
    {
        public LithiumCompilerException(string messageFormat, params string[] args)
            : base(messageFormat, args)
        {

        }
    }
}
