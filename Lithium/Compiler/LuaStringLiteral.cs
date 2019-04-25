using Irony;
using Irony.Ast;
using Irony.Parsing;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;

namespace Lithium.Compiler
{

    public class LuaStringLiteral : Terminal
    {
        public LuaStringLiteral(string name)
            : base(name)
        {
            Escapes = GetDefaultEscapes();
            SetFlag(TermFlags.IsLiteral);
            SetFlag(TermFlags.IsMultiline);
            EditorInfo = new TokenEditorInfo(TokenType.String, TokenColor.String, TokenTriggers.MatchBraces);
        }

        public static EscapeTable GetDefaultEscapes()
        {
            EscapeTable escapes = new EscapeTable();
            escapes.Add('a', '\u0007');
            escapes.Add('b', '\b');
            escapes.Add('t', '\t');
            escapes.Add('n', '\n');
            escapes.Add('v', '\v');
            escapes.Add('f', '\f');
            escapes.Add('r', '\r');
            escapes.Add('"', '"');
            escapes.Add('\'', '\'');
            escapes.Add('\\', '\\');
            escapes.Add('[', '[');
            escapes.Add(']', ']');
            return escapes;
        }

        protected abstract class StringMatcher
        {
            public StringMatcher(StringOptions flags)
            {
                Flags = flags;
            }

            public readonly StringOptions Flags;
            public abstract string TryGetInitiator(string text, int startAt, bool ignoreCase);
            public abstract string TryGetTerminator(string initiator, bool ignoreCase);

            protected bool StartsWith(string text, string source, int startAt, bool ignoreCase)
            {
                if (source.Length - startAt < text.Length) return false;

                for (int i = startAt, j = 0; j < text.Length; i++, j++)
                {
                    if (!ignoreCase && source[i] != text[j]) return false;
                    if (ignoreCase && char.ToLower(source[i]) != char.ToLower(text[j])) return false;
                }
                return true;
            }
        }

        protected class StandardStringMatcher : StringMatcher
        {
            readonly string Start, End;
            public StandardStringMatcher(string start, string end, StringOptions flags)
                : base(flags)
            {
                Start = start;
                End = end;
            }

            public override string TryGetInitiator(string text, int startAt, bool ignoreCase)
            {
                if (StartsWith(Start, text, startAt, ignoreCase))
                    return Start;
                return null;
            }

            public override string TryGetTerminator(string initiator, bool ignoreCase)
            {
                return End;
            }
        }

        protected class RegexStringMatcher : StringMatcher
        {
            readonly Regex Start;
            readonly string StartShortcut;
            readonly string Format;
            public RegexStringMatcher(Regex start, string shortcut, string endFormat, StringOptions flags)
                : base(flags)
            {
                Start = start;
                StartShortcut = shortcut;
                Format = endFormat;
            }

            public override string TryGetInitiator(string text, int startPos, bool ignoreCase)
            {
                if (!StartsWith(StartShortcut, text, startPos, ignoreCase))
                    return null;

                var match = Start.Match(text, startPos);
                if (!match.Success || match.Index != startPos)
                    return null;
                return match.Value;
            }

            public override string TryGetTerminator(string initiator, bool ignoreCase)
            {
                return Start.Replace(initiator, Format);
            }
        }


        private List<StringMatcher> _matchers = new List<StringMatcher>();
        public EscapeTable Escapes = new EscapeTable();

        public override void Init(GrammarData grammarData)
        {
            base.Init(grammarData);



            _matchers.Add(new StandardStringMatcher("'", "'", StringOptions.AllowsAllEscapes));
            _matchers.Add(new StandardStringMatcher("\"", "\"", StringOptions.AllowsAllEscapes));
            _matchers.Add(new RegexStringMatcher(new Regex(@"\[(=*)\[", RegexOptions.Compiled), "[", @"]$1]", StringOptions.AllowsLineBreak | StringOptions.NoEscapes));


        }



        private bool IsEndQuoteEscaped(string text, int quotePosition)
        {
            bool escaped = false;
            int p = quotePosition - 1;
            while (p > 0 && text[p] == '\\')
            {
                escaped = !escaped;
                p--;
            }
            return escaped;
        }

        public override Token TryMatch(ParsingContext context, ISourceStream source)
        {
            foreach (var m in _matchers)
            {
                var backupOffset = source.PreviewPosition;

                var start = m.TryGetInitiator(source.Text, source.PreviewPosition, !context.Language.Grammar.CaseSensitive);
                if (start != null)
                {
                    //Locate the end terminal
                    var end = m.TryGetTerminator(start, !context.Language.Grammar.CaseSensitive);

                    if (end == null)
                        continue;

                    source.PreviewPosition += start.Length;

                    var startPos = source.PreviewPosition;
                    var endPos = -2;
                    while (endPos == -2)
                    {
                        endPos = source.Text.IndexOf(end, source.PreviewPosition);

                        if ((m.Flags & StringOptions.NoEscapes) != 0)
                            continue;

                        if (IsEndQuoteEscaped(source.Text, endPos))
                        {
                            source.PreviewPosition = endPos + end.Length;
                            endPos = -2;
                        }

                    }
                    if (endPos == -1) continue;

                    source.PreviewPosition = endPos + end.Length;
                    var tokenText = source.Text.Substring(startPos, endPos - startPos);

                    if ((m.Flags & StringOptions.NoEscapes) == 0)
                        tokenText = ConvertValue(context, tokenText, m);

                    return source.CreateToken(this.OutputTerminal, tokenText);
                }
            }
            return null;
        }



        //Extract the string content from lexeme, adjusts the escaped and double-end symbols
        protected string ConvertValue(ParsingContext context, string segment, StringMatcher details)
        {
            string value = segment.Replace("\r\n", "\n");
            bool escapeEnabled = (details.Flags & StringOptions.NoEscapes) == 0;
            //Fix all escapes
            if (escapeEnabled && value.IndexOf('\\') >= 0)
            {
                string[] arr = value.Split('\\');
                bool ignoreNext = false;
                //we skip the 0 element as it is not preceeded by "\"
                for (int i = 1; i < arr.Length; i++)
                {
                    if (ignoreNext)
                    {
                        ignoreNext = false;
                        continue;
                    }
                    string s = arr[i];
                    if (string.IsNullOrEmpty(s))
                    {
                        //it is "\\" - escaped escape symbol. 
                        arr[i] = @"\";
                        ignoreNext = true;
                        continue;
                    }
                    else if (s[0] == 'z')
                    {
                        //Lua's whitespace gobbler
                        arr[i] = arr[i].Substring(1).TrimStart(' ', '\t', '\n', '\r');
                        continue;
                    }
                    //The char is being escaped is the first one; replace it with char in Escapes table
                    char first = s[0];
                    char newFirst;
                    if (Escapes.TryGetValue(first, out newFirst))
                        arr[i] = newFirst + s.Substring(1);
                    else
                    {
                        arr[i] = HandleSpecialEscape(context, arr[i], details);
                    }//else
                }//for i
                value = string.Join(string.Empty, arr);
            }// if EscapeEnabled 
            else
            {
                if (value.Length > 0 && value[0] == '\n')
                    return value.Substring(1);
            }

            return value;
        }

        //Should support:  \Udddddddd, \udddd, \xdddd, \N{name}, \0, \ddd (ascii),  
        protected string HandleSpecialEscape(ParsingContext context, string segment, StringMatcher details)
        {
            if (string.IsNullOrEmpty(segment)) return string.Empty;
            int len, p; string digits; char ch; string result;
            char first = segment[0];
            switch (first)
            {
                case 'u':
                case 'U':
                    if ((details.Flags & StringOptions.AllowsUEscapes) != 0)
                    {
                        len = (first == 'u' ? 4 : 8);
                        if (segment.Length < len + 1)
                        {
                            context.AddParserError(Resources.ErrBadUnEscape, segment.Substring(len + 1), len);
                            return null;
                        }
                        digits = segment.Substring(1, len);
                        ch = (char)Convert.ToUInt32(digits, 16);
                        result = ch + segment.Substring(len + 1);
                        return result;
                    }//if
                    break;
                case 'x':
                    if ((details.Flags & StringOptions.AllowsXEscapes) != 0)
                    {
                        //x-escape allows variable number of digits, from one to 4; let's count them
                        p = 1; //current position
                        while (p < 5 && p < segment.Length)
                        {
                            if (Strings.HexDigits.IndexOf(segment[p]) < 0) break;
                            p++;
                        }
                        //p now point to char right after the last digit
                        if (p <= 1)
                        {
                            context.AddParserError(Resources.ErrBadXEscape); // @"Invalid \x escape, at least one digit expected.";
                            return null;
                        }
                        digits = segment.Substring(1, p - 1);
                        ch = (char)Convert.ToUInt32(digits, 16);
                        result = ch + segment.Substring(p);
                        return result;
                    }//if
                    break;
                case '0':
                case '1':
                case '2':
                case '3':
                case '4':
                case '5':
                case '6':
                case '7':
                case '8':
                case '9':
                    if ((details.Flags & StringOptions.NoEscapes) == 0)
                    {
                        //octal escape allows variable number of digits, from one to 3; let's count them
                        p = 0; //current position
                        while (p < 3 && p < segment.Length)
                        {
                            if (Strings.DecimalDigits.IndexOf(segment[p]) < 0) break;
                            p++;
                        }
                        //p now point to char right after the last digit
                        digits = segment.Substring(0, p);
                        ch = (char)Convert.ToUInt32(digits);
                        result = ch + segment.Substring(p);
                        return result;
                    }//if
                    break;
            }//switch
            context.AddParserError(Resources.ErrInvEscape, segment); //"Invalid escape sequence: \{0}"
            return null;
        }//method

        public override IList<string> GetFirsts()
        {
            return new[] { "'", "\"", "[" };
        }
    }

}
