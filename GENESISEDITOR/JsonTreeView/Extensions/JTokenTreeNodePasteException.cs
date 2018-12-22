using System;

namespace JsonTreeView.Extensions
{
    public class JTokenTreeNodePasteException : AggregateException
    {
        public JTokenTreeNodePasteException(Exception sourceException)
            : base(sourceException)
        {
        }
    }
}
