using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

using System.Linq.Expressions;
using Lithium.Exceptions;
using Irony.Parsing;

namespace Lithium.Compiler
{
    abstract class CompilationScope : IDisposable
    {
        protected CompilationScope(Generator generator, SourceSpan span)
        {
            Generator = generator;
            Span = span;
            ParentScope = Generator.CurrentScope;

            Environment = Expression.Parameter(typeof(object), "E");

            Generator.CurrentScope = this;
        }

        /// <summary>
        /// Gets the reference to the current compilation <see cref="Generator"/>
        /// </summary>
        public Generator Generator
        { get; private set; }

        /// <summary>
        /// Gets the <see cref="SourceSpan"/> in which this block is defined
        /// </summary>
        public SourceSpan Span
        { get; private set; }

        /// <summary>
        /// Gets the parent compilation scope, or <c>null</c> if this is the root scope
        /// </summary>
        public CompilationScope ParentScope
        { get; protected set; }

        /// <summary>
        /// Gets the reference to this scope's environment variable
        /// </summary>
        public ParameterExpression Environment
        { get; private set; }

        /// <summary>
        /// Gets all available variables within this scope, including the environment variable
        /// </summary>
        public IEnumerable<ParameterExpression> Variables
        {
            get
            {
                yield return Environment;

                foreach (var v in LocalVariables)
                    yield return v;
            }
        }

        /// <summary>
        /// Gets all available local variables within this scope
        /// </summary>
        public virtual IEnumerable<ParameterExpression> LocalVariables
        {
            get
            {
                yield break;
            }
        }

        /// <summary>
        /// Registers a new local variable in the nearest supporting scope
        /// </summary>
        /// <param name="identifier">The identifier to be used for accessing this variable</param>
        public virtual ParameterExpression RegisterLocal(string identifier)
        {
            if (ParentScope == null)
                throw new LithiumCompilerException("Cannot register local variable outside of a valid block scope.");

            return ParentScope.RegisterLocal(identifier);
        }

        /// <summary>
        /// Gets am expression for a local variable or global environment object which
        /// will return the value for a variable when executed.
        /// </summary>
        /// <param name="identifier">The variable's identifier</param>
        public virtual Expression GetVariableForRead(string identifier)
        {
            if (!LocalVariables.Any() && ParentScope == null)
                return Expressions.MakeGlobalVariableAccessor(Environment, identifier);

            //Navigate up the scope tree to see if we can find a local variable with the given identifier
            return LocalVariables.SingleOrDefault(x => x.Name == identifier) ?? ParentScope.GetVariableForRead(identifier);
        }

        /// <summary>
        /// Gets an expression which, when executed, will write the given value expression
        /// to any local variable with the given identifier, and if no local variable exists
        /// will write to the global scope.
        /// </summary>
        /// <param name="identifier">The variable's identifier</param>
        /// <param name="value">An expression for the value to be assigned to the variable</param>
        public virtual Expression WriteVariable(string identifier, Expression value)
        {
            if (!LocalVariables.Any() && ParentScope == null)
                return Expressions.MakeGlobalVariableWriter(Environment, identifier, value.EnsureObject());

            if (LocalVariables.Any(x => x.Name == identifier))
                return Expression.Assign(LocalVariables.Single(x => x.Name == identifier), value.EnsureObject());
            else
                return ParentScope.WriteVariable(identifier, value);
        }
        
        /// <summary>
        /// Gets the <see cref="LabelTarget"/> used to exit the current function
        /// </summary>
        public virtual LabelTarget GetReturnLabel()
        {
            if (ParentScope == null)
                throw new LithiumCompilerException("Return expression did not exist within a valid function scope.");

            return ParentScope.GetReturnLabel();
        }

        /// <summary>
        /// Gets the <see cref="LabelTarget"/> used to exit the current loop
        /// </summary>
        public virtual LabelTarget GetExitLabel()
        {
            if (ParentScope == null)
                throw new LithiumCompilerException("Break expression did not exist within a valid loop.");

            return ParentScope.GetExitLabel();
        }

        /// <summary>
        /// Registers a Goto label which can be used to execute a jump within the code.
        /// </summary>
        /// <param name="labelName">The name of the label to register</param>
        public virtual Expression RegisterJumpTarget(string labelName)
        {
            if (ParentScope == null)
                throw new LithiumCompilerException("Cannot register jump label in current scope or any of its parents.");

            return ParentScope.RegisterJumpTarget(labelName);
        }

        /// <summary>
        /// Gets the <see cref="LabelTarget"/> used to complete a "goto {label}" operation.
        /// </summary>
        /// <param name="labelName">The name of the label to jump to</param>
        public virtual LabelTarget GetJumpLabel(string labelName)
        {
            if (ParentScope == null)
                throw new LithiumCompilerException("Couldn't locate label {0} to jump to.", labelName);

            return ParentScope.GetJumpLabel(labelName);
        }

        public void Dispose()
        {
            Generator.CurrentScope = ParentScope;
        }

        public override string ToString()
        {
            return string.Format("{0} @ {1}", this.GetType().Name, Span.Location.ToUiString());
        }
    }

    class BlockCompilationScope : CompilationScope
    {
        public BlockCompilationScope(Generator generator, SourceSpan span)
            : base(generator, span)
        {
        }

        List<ParameterExpression> _registeredLocals = new List<ParameterExpression>();

        Dictionary<string, LabelTarget> _gotoLabels = new Dictionary<string, LabelTarget>();

        /// <inheritdoc/>
        public override IEnumerable<ParameterExpression> LocalVariables
        {
            get
            {
                return _registeredLocals;
            }
        }

        /// <inheritdoc/>
        public override ParameterExpression RegisterLocal(string identifier)
        {
            //Don't allow us to register more than 1 local with the same ID
            if (_registeredLocals.Any(x => x.Name == identifier))
                return _registeredLocals.Single(x => x.Name == identifier);

            var newVariable = Expression.Variable(typeof(object), identifier);
            _registeredLocals.Add(newVariable);
            return newVariable;
        }

        /// <inheritdoc/>
        public override Expression RegisterJumpTarget(string labelName)
        {
            var target = Expression.Label();
            _gotoLabels.Add(labelName, target);
            return Expression.Label(target);
        }

        /// <inheritdoc/>
        public override LabelTarget GetJumpLabel(string labelName)
        {
            if (_gotoLabels.ContainsKey(labelName))
                return _gotoLabels[labelName];

            return base.GetJumpLabel(labelName);
        }
    }

    class FunctionScope : CompilationScope
    {
        public FunctionScope(Generator generator, SourceSpan span, string name)
            : base(generator, span)
        {
            Name = name;
            ReturnLabel = Expression.Label();
        }

        public string Name
        { get; private set; }

        public LabelTarget ReturnLabel
        { get; private set; }

        /// <inheritdoc/>
        public override LabelTarget GetReturnLabel()
        {
            return ReturnLabel;
        }

        public override IEnumerable<ParameterExpression> LocalVariables
        {
            get
            {
                return _parameters;
            }
        }

        List<ParameterExpression> _parameters = new List<ParameterExpression>();

        public virtual ParameterExpression RegisterParameter(string name)
        {
            var p = Expression.Parameter(typeof(object), name);
            _parameters.Add(p);
            return p;
        }

        public override string ToString()
        {
            return string.Format("FunctionScope {0} @ {1}", Name, Span.Location.ToUiString());
        }
    }

    class LoopScope : CompilationScope
    {
        public LoopScope(Generator generator, SourceSpan span)
            : base(generator, span)
        {
            ExitLabel = Expression.Label();
        }

        public LabelTarget ExitLabel
        { get; private set; }

        public override LabelTarget GetExitLabel()
        {
            return ExitLabel;
        }
    }

    class IteratorCompilationScope : LoopScope
    {
        public IteratorCompilationScope(Generator generator, SourceSpan span, string loopVariableName)
            : base(generator, span)
        {
            LoopVariable = Expression.Variable(typeof(double), loopVariableName);
        }

        public ParameterExpression LoopVariable
        { get; private set; }

        public override IEnumerable<ParameterExpression> LocalVariables
        {
            get
            {
                yield return LoopVariable;
            }
        }
    }
}
