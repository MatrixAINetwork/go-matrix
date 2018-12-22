using System;

namespace JsonTreeView
{
    public class AfterSelectEventArgs : EventArgs
    {
        public AfterSelectEventArgs(string typeName, string jTokenTypeName, Func<string> getJsonString)
        {
            TypeName = typeName;
            JTokenTypeName = jTokenTypeName;
            GetJsonString = getJsonString;
        }

        public string TypeName { get; private set; }
        public string JTokenTypeName { get; }
        public Func<string> GetJsonString { get; }
    }
}