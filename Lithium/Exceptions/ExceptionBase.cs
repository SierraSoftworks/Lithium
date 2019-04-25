using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Lithium.Exceptions
{
    public abstract class LithiumException : Exception
    {
        public LithiumException(string messageFormat, params string[] args)
            : base(string.Format(messageFormat, args))
        {

        }
    }
}
