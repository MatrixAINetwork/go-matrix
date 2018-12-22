using System.IO;
using Newtonsoft.Json.Linq;

namespace JsonTreeView
{
    internal sealed class JTokenRoot
    {
        #region >> Fields

        private JToken jTokenValue;

        #endregion

        #region >> Properties

        /// <summary>
        /// Root <see cref="JToken"/> node.
        /// </summary>
        public JToken JTokenValue
        {
            get { return jTokenValue; }
            set { jTokenValue = value; }
        }

        #endregion

        #region >> Constructors

        /// <summary>
        /// Constructor using an existing stream to populate the instance.
        /// </summary>
        /// <param name="jsonStream">Source stream.</param>
        public JTokenRoot(Stream jsonStream)
        {
            Load(jsonStream);
        }

        /// <summary>
        /// Constructor using an existing json string to populate the instance.
        /// </summary>
        /// <param name="jsonString">Source string.</param>
        public JTokenRoot(string jsonString)
        {
            Load(jsonString);
        }

        /// <summary>
        /// Constructor using an existing json string to populate the instance.
        /// </summary>
        /// <param name="jToken">Source <see cref="JToken"/>.</param>
        public JTokenRoot(JToken jToken)
        {
            Load(jToken);
        }

        #endregion

        /// <summary>
        /// Initialize using an existing stream to populate the instance.
        /// </summary>
        /// <param name="jsonStream">Source stream.</param>
        public void Load(Stream jsonStream)
        {
            using (var streamReader = new StreamReader(jsonStream))
            {
                Load(streamReader.ReadToEnd());
            }
        }

        /// <summary>
        /// Initialize using an existing json string to populate the instance.
        /// </summary>
        /// <param name="jsonString">Source string.</param>
        public void Load(string jsonString)
        {
            Load(JToken.Parse(jsonString));
        }

        /// <summary>
        /// Initialize using an existing json string to populate the instance.
        /// </summary>
        /// <param name="jToken">Source <see cref="JToken"/>.</param>
        public void Load(JToken jToken)
        {
            jTokenValue = jToken;
        }

        /// <summary>
        /// Save the enclosed <see cref="JToken"/> in an existing stream.
        /// </summary>
        /// <param name="jsonStream">Target stream.</param>
        public void Save(Stream jsonStream)
        {
            if (jTokenValue == null)
            {
                return;
            }

            using (var streamWriter = new StreamWriter(jsonStream))
            {
                streamWriter.Write(jTokenValue.ToString());
            }
        }
    }
}
