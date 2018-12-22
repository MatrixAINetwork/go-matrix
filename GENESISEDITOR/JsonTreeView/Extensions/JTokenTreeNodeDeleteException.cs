using System;

namespace JsonTreeView.Extensions
{
    public class JTokenTreeNodeDeleteException : AggregateException
    {
        public JTokenTreeNodeDeleteException(Exception sourceException)
            : base(sourceException)
        {
        }
    }
}
