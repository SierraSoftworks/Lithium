using Irony;
using Irony.Parsing;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;

namespace Lithium.Compiler
{
    public class LuaLongCommentTerminal : Terminal
    {
        public LuaLongCommentTerminal(string name)
            : base(name, TokenCategory.Comment)
        {
            Priority = TerminalPriority.High; //assign max priority
        }

        #region overrides
        Regex counterRegex = new Regex(@"--\[(=*)\[", RegexOptions.Compiled);
        public override void Init(GrammarData grammarData)
        {
            base.Init(grammarData);

            if (this.EditorInfo == null)
            {
                TokenType ttype = TokenType.Comment;
                this.EditorInfo = new TokenEditorInfo(ttype, TokenColor.Comment, TokenTriggers.None);
            }
        }

        public override Token TryMatch(ParsingContext context, ISourceStream source)
        {
            Token result;
            int equalsCount = 0;
            if (context.VsLineScanState.Value != 0)
            {
                // we are continuing in line mode - restore internal env (none in this case)
                context.VsLineScanState.Value = 0;
            }
            else
            {
                //we are starting from scratch
                if (!BeginMatch(context, source, out equalsCount)) return null;
            }
            result = CompleteMatch(context, source, equalsCount);
            if (result != null) return result;
            if (context.Mode == ParseMode.VsLineScan)
                return CreateIncompleteToken(context, source);
            return context.CreateErrorToken(Resources.ErrUnclosedComment);
        }

        private Token CreateIncompleteToken(ParsingContext context, ISourceStream source)
        {
            source.PreviewPosition = source.Text.Length;
            Token result = source.CreateToken(this.OutputTerminal);
            result.Flags |= TokenFlags.IsIncomplete;
            context.VsLineScanState.TerminalIndex = this.MultilineIndex;
            return result;
        }

        private bool BeginMatch(ParsingContext context, ISourceStream source, out int numberOfEquals)
        {
            numberOfEquals = 0;

            //Check starting symbol
            if (!source.MatchSymbol("--[")) return false;

            //Do our regex magic
            var match = counterRegex.Match(source.Text, source.PreviewPosition);
            numberOfEquals = match.Groups[1].Length;

            source.PreviewPosition += match.Length;
            return true;
        }
        private Token CompleteMatch(ParsingContext context, ISourceStream source, int numberOfEquals)
        {
            //Find end symbol

            string endSymbol = "]";
            while (numberOfEquals-- > 0) endSymbol += "=";
            endSymbol += "]";

            int firstCharPos = source.Text.IndexOf(endSymbol, source.PreviewPosition);

            if (firstCharPos < 0)
            {
                source.PreviewPosition = source.Text.Length;
                return null; //indicating error
            }

            source.PreviewPosition = firstCharPos + endSymbol.Length;
            return source.CreateToken(this.OutputTerminal);

        }//method

        public override IList<string> GetFirsts()
        {
            return new string[] { "--[" };
        }
        #endregion
    }//CommentTerminal class
}
