using System;

namespace JsonTreeView
{
    public class WrongJsonStreamException : Exception
    {
        public WrongJsonStreamException(string message, Exception innerException) : base(message, innerException)
        {
        }
    }
}